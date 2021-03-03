package discovery

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

func NewEmitter(address string) *p2pUDPEmitter {
	return &p2pUDPEmitter{
		address: address,
	}
}

type p2pUDPEmitter struct {
	address string

	stopInit chan struct{}
	stopDone chan struct{}
}

func (e p2pUDPEmitter) Run() {
	addr, err := net.ResolveUDPAddr("udp", e.address)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)

	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-e.stopInit:
			break LOOP
		case <-ticker.C:
			peer := Peer{Address: GetOutboundIP().String()}
			bytes, err := json.Marshal(peer)
			if err != nil {
				log.Fatal(err)
			}
			_, _ = c.Write(bytes)
		}
	}

	_ = c.Close()
}

func (e p2pUDPEmitter) Shutdown() {
	e.stopInit <- struct{}{}
	<-e.stopDone
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
