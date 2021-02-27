package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"

	"hunched-dog/internal"
	"hunched-dog/pkg/shutdowner"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/.hunched-dog")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func main() {

	targetDirectory := viper.GetString("target")
	targetDirectory = ReplacePath(targetDirectory)

	log.Println("INFO", "create directory", targetDirectory)
	err := os.MkdirAll(targetDirectory, os.ModePerm)
	if err != nil {
		log.Fatalln("ERR", err.Error())
	}

	go func() {
		fs := http.FileServer(http.Dir(targetDirectory))
		for _, port := range allowedFilePorts {
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), fs); err != nil {
				log.Println("WARN", "can't listen port", port, ":", err.Error())
			}
		}
	}()

	go func() {
		http.HandleFunc("/registry", func(w http.ResponseWriter, req *http.Request) {
			reg, err := internal.GetLocal(targetDirectory)
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
						continue
					}

					bytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println("ERR", err.Error())
						continue
					}
					remoteReg := internal.Registry{}
					err = json.Unmarshal(bytes, &remoteReg)
					if err != nil {
						log.Println("ERR", err.Error())
						continue
					}

					localReg, err := internal.GetLocal(targetDirectory)
					if err != nil {
						log.Println("ERR", err.Error())
						continue
					}

					dirs := internal.DiffDirs(remoteReg)
					for _, d := range dirs {
						log.Println("INFO", "create directory", d)
						err = os.MkdirAll(targetDirectory+"/"+d, os.ModePerm)
						if err != nil {
							log.Println("ERR", err.Error())
							continue
						}
					}

					diffReg := internal.DiffFiles(localReg, remoteReg)

					for _, filePort := range allowedFilePorts {
						for _, filename := range diffReg {
							log.Println("INFO", "download file", filename)

							// Get the data
							resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", h, filePort, filename))
							if err != nil {
								log.Println("ERR", err.Error())
								continue
							}
							defer resp.Body.Close()

							out, err := os.Create(targetDirectory + "/" + filename)
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
