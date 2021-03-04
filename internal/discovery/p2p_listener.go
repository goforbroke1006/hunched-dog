package discovery

import (
	"log"
	"net"
)

const (
	maxDatagramSize = 8192
)

func NewListener(address string) *p2pUDPListener {
	return &p2pUDPListener{
		address: address,

		currentPeer: GetOutboundIP().String(),
		peers:       make(chan string, 12),

		stopInit: make(chan struct{}),
		stopDone: make(chan struct{}),
	}
}

type p2pUDPListener struct {
	address string

	currentPeer string
	peers       chan string

	stopInit chan struct{}
	stopDone chan struct{}
}

func (l *p2pUDPListener) Peers() chan string {
	return l.peers
}

func (l *p2pUDPListener) Run() {
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

			peerIP := src.IP.String()

			if l.currentPeer == peerIP { // skip current IP
				continue
			}

			log.Println(n, "bytes read from", peerIP)
			//log.Println(hex.Dump(b[:n]))

			l.peers <- peerIP
		}
	}

	_ = conn.Close()
	l.stopDone <- struct{}{}
}

func (l *p2pUDPListener) Shutdown() {
	l.stopInit <- struct{}{}
	<-l.stopDone
}
