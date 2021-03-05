package discovery

import (
	"log"
	"net"
	"time"
)

func NewEmitter(address string) *p2pUDPEmitter {
	return &p2pUDPEmitter{
		address: address,

		stopInit: make(chan struct{}),
		stopDone: make(chan struct{}),
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
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(10 * time.Second)

LOOP:
	for {
		select {
		case <-e.stopInit:
			break LOOP
		case <-ticker.C:
			message := []byte("hunched-dog")

			err = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				log.Fatal(err)
			}

			_, err = c.Write(message)
			if err != nil {
				log.Fatal(err)
			}
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
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
