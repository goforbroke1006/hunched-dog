package discovery

import (
	"encoding/json"
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

func NewListener(address string) *p2pUDPListener {
	return &p2pUDPListener{
		address: address,
	}
}

type p2pUDPListener struct {
	address string

	stopInit chan struct{}
	stopDone chan struct{}
}

func (l p2pUDPListener) Run() {
	udpAddr, err := net.ResolveUDPAddr("udp", l.address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	_ = conn.SetReadBuffer(maxDatagramSize)

LOOP:
	for {
		select {
		case <-l.stopInit:
			break LOOP
		default:
			b := make([]byte, maxDatagramSize)
			n, src, err := conn.ReadFromUDP(b)
			if err != nil {
				log.Fatal("ReadFromUDP failed:", err)
			}
			peer := Peer{}
			err = json.Unmarshal(b, &peer)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("find new peek", peer.Address)
			log.Println(n, "bytes read from", src)

			//log.Println(hex.Dump(b[:n]))
		}
	}

	_ = conn.Close()
	l.stopDone <- struct{}{}
}

func (l p2pUDPListener) Shutdown() {
	l.stopInit <- struct{}{}
	<-l.stopDone
}
