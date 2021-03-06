package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"hunched-dog/internal/discovery"
	"hunched-dog/internal/downloader"
	"hunched-dog/internal/registry"
	"hunched-dog/pkg/shutdowner"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("$HOME/.hunched-dog")
	viper.AddConfigPath(".")

	// hotfix for windows
	if userHomeDir, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(userHomeDir, ".hunched-dog"))
	}

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

	currentIP := discovery.GetOutboundIP().String()
	log.Println("INFO", "current host", currentIP)

	go func() {
		fs := http.FileServer(http.Dir(targetDirectory))
		for _, port := range allowedFilePorts {
			log.Println("INFO", "try to listen port", port, "for fs endpoint")
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", currentIP, port), fs); err != nil {
				log.Println("WARN", "can't listen port", port, ":", err.Error())
			}
		}
	}()

	go func() {
		http.HandleFunc("/registry", func(w http.ResponseWriter, req *http.Request) {
			reg, err := registry.GetLocal(targetDirectory)
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
			log.Println("INFO", "try to listen port", port, "for http endpoint")
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", currentIP, port), nil); err != nil {
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
