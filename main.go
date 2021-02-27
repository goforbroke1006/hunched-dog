package main

import (
	"encoding/json"
	"fmt"
	"hunched-dog/internal"
	"hunched-dog/pkg/shutdowner"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

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
			reg, err := internal.GetLocal(directory)
			if err != nil {
				log.Fatal(err)
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
		for {
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
					remoteReg := internal.Registry{}
					err = json.Unmarshal(bytes, &remoteReg)
					if err != nil {
						log.Println("ERR", err.Error())
					}

					localReg, err := internal.GetLocal(directory)
					if err != nil {
						log.Println("ERR", err.Error())
					}
					diffReg := internal.Diff(localReg, remoteReg)

					for _, filePort := range allowedFilePorts {
						for _, filename := range diffReg {
							log.Println("INFO", "download file ", filename)

							// Get the data
							resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", h, filePort, filename))
							if err != nil {
								log.Println("ERR", err.Error())
								continue
							}
							defer resp.Body.Close()

							out, err := os.Create(directory + "/" + filename)
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

						break // TODO: fix port availability check
					}
				}
			}

			time.Sleep(30 * time.Second)
		}
	}()

	shutdowner.WaitTermination()
}
