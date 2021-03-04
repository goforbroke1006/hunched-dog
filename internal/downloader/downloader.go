package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"hunched-dog/internal"
)

func New(
	target string,
	peers chan string,
	httpPorts, filePorts []int,
) *filesDownloader {
	return &filesDownloader{
		target:    target,
		peers:     peers,
		httpPorts: httpPorts,
		filePorts: filePorts,

		activePeers: make(map[string]struct{}, 12),

		stopInit: make(chan struct{}),
		stopDone: make(chan struct{}),
	}
}

type filesDownloader struct {
	target    string
	peers     chan string
	httpPorts []int
	filePorts []int

	activePeers   map[string]struct{}
	activePeersMx sync.RWMutex

	stopInit chan struct{}
	stopDone chan struct{}
}

func (d *filesDownloader) Run() {
LOOP:
	for {
		select {
		case <-d.stopInit:
			break LOOP
		case peerIP := <-d.peers:
			d.onNewPeer(peerIP)
		}
	}
	d.stopDone <- struct{}{}
}

func (d *filesDownloader) onNewPeer(peerIP string) {
	d.activePeersMx.RLock()
	if _, ok := d.activePeers[peerIP]; ok {
		return
	}
	d.activePeersMx.RUnlock()

	d.activePeersMx.Lock()
	d.activePeers[peerIP] = struct{}{}
	d.activePeersMx.Unlock()

	for _, port := range d.httpPorts {
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/registry", peerIP, port))
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

		localReg, err := internal.GetLocal(d.target)
		if err != nil {
			log.Println("ERR", err.Error())
			continue
		}

		dirs := internal.DiffDirs(remoteReg)
		for _, dir := range dirs {
			log.Println("INFO", "create directory", dir)
			err = os.MkdirAll(path.Join(d.target, dir), os.ModePerm)
			if err != nil {
				log.Println("ERR", err.Error())
				continue
			}
		}

		diffReg := internal.DiffFiles(localReg, remoteReg)

		for _, filePort := range d.filePorts {
			for _, metaFile := range diffReg {
				log.Println("INFO", "download file", metaFile.Filename)

				// Get the data
				resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", peerIP, filePort, metaFile.Filename))
				if err != nil {
					log.Println("ERR", err.Error())
					continue
				}
				defer resp.Body.Close()

				asbFilename := path.Join(d.target, metaFile.Filename)
				out, err := os.Create(asbFilename)
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

				modificationDate := time.Unix(metaFile.UpdatedAt, 0)
				err = os.Chtimes(asbFilename, modificationDate, modificationDate)
				if err != nil {
					log.Println("ERR", err.Error())
					continue
				}

			}

			break // TODO: fix port availability check
		}

		break
	}

	d.activePeersMx.Lock()
	delete(d.activePeers, peerIP)
	d.activePeersMx.Unlock()
}

func (d *filesDownloader) Shutdown() {
	d.stopInit <- struct{}{}
	<-d.stopDone
}
