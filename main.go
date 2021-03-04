package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"hunched-dog/internal"
	"hunched-dog/internal/discovery"
	"hunched-dog/internal/downloader"
	"hunched-dog/pkg/shutdowner"
	"log"
	"net/http"
	"os"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/.hunched-dog")
	viper.AddConfigPath("%UserProfile%/.hunched-dog")
	viper.AddConfigPath("$UserProfile/.hunched-dog")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func main() {

	multicast := viper.GetString("multicast")

	p2pUDPListener := discovery.NewListener(multicast)
	go p2pUDPListener.Run()
	defer p2pUDPListener.Shutdown()

	p2pUDPEmitter := discovery.NewEmitter(multicast)
	go p2pUDPEmitter.Run()
	defer p2pUDPEmitter.Shutdown()

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

		for _, port := range allowedHttpPorts {
			if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil); err != nil {
				log.Println("WARN", "can't listen port", port, ":", err.Error())
			}
		}
	}()

	filesDownloader := downloader.New(targetDirectory, p2pUDPListener.Peers(),
		allowedHttpPorts, allowedFilePorts)
	go filesDownloader.Run()
	defer filesDownloader.Shutdown()

	shutdowner.WaitTermination()
}
