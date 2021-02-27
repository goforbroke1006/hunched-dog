package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"hunched-dog/pkg/shutdowner"
)

type MetaFile struct {
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	UpdatedAt int64  `json:"updated_at"`
}
type Registry []MetaFile

func main() {

	go func() {
		fs := http.FileServer(http.Dir(directory))
		for _, port := range allowedFilePorts {
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), fs); err != nil {
				log.Println("WARN", "can't listen port", port, ":", err.Error())
			}
		}
	}()

	go func() {
		http.HandleFunc("/registry", func(w http.ResponseWriter, req *http.Request) {
			reg := Registry{}

			err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				reg = append(reg, MetaFile{
					Filename:  path,
					Size:      info.Size(),
					UpdatedAt: info.ModTime().Unix(),
				})
				return nil
			})
			if err != nil {
				log.Println(err)
			}

			bytes, err := json.MarshalIndent(reg, "", "  ")
			if err != nil {
				log.Fatal(err)
			}

			w.WriteHeader(200)
			_, err = w.Write(bytes)
			if err != nil {
				log.Fatal(err)
			}
		})

		for _, port := range allowedPorts {
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil); err != nil {
				log.Println("WARN", "can't listen port", port, ":", err.Error())
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-ticker.C:
				for _, h := range hosts {
					for _, p := range allowedPorts {
						resp, err := http.Get(fmt.Sprintf("http://%s:%d/registry", h, p))
						if err != nil {
							log.Println("ERR", err.Error())
						}

						bytes, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							log.Println("ERR", err.Error())
						}
						remoteReg := Registry{}
						err = json.Unmarshal(bytes, &remoteReg)
						if err != nil {
							log.Println("ERR", err.Error())
						}

						diffReg := Registry{}
						diffReg = remoteReg

						for _, filePort := range allowedFilePorts {
							for _, file := range diffReg {

								// Get the data
								resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", h, filePort, file.Filename))
								if err != nil {
									log.Println("ERR", err.Error())
									continue
								}
								defer resp.Body.Close()

								out, err := os.Create(directory + "/" + file.Filename)
								if err != nil {
									log.Println("ERR", err.Error())
									continue
								}
								defer out.Close()

								// Writer the body to file
								_, err = io.Copy(out, resp.Body)
								if err != nil {
									log.Println("ERR", err.Error())
									continue
								}
							}
						}
					}
				}
			}
		}
	}()

	shutdowner.WaitTermination()
}
