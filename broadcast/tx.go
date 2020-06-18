package broadcast

import (
	"fmt"
	"log"
	"net"
	"sync"

	"realizr.io/gome/model"
)

const MAX_MSG_SIZE = 2048000
const CMD_PORT = "127.0.0.1:7456"

func Send(logEntry model.LogEntry, subscribers []string) {
	dat := model.Encode(logEntry)

	for _, element := range subscribers {
		go tx(element, dat)
	}
}

func tx(address string, msg []byte) {
	conn, _ := net.Dial("udp", address)
	defer conn.Close()
	conn.Write(msg)
	log.Printf("tx=>%s\n", address)
}

func Listen(wg *sync.WaitGroup) {
	defer wg.Done()

	addr, _ := net.ResolveUDPAddr("udp", CMD_PORT)
	sock, _ := net.ListenUDP("udp", addr)
	defer sock.Close()

	i := 0
	buf := make([]byte, MAX_MSG_SIZE)
	for {
		rlen, remoAddr, err := sock.ReadFromUDP(buf)
		i += rlen
		if err != nil {
			fmt.Println(err)
		} else {
			le := model.ReceiptLogEntry(buf[0:rlen])
			le.Origin = remoAddr.IP.String()
			log.Printf("JSON %s\n", le.ToJSON())
		}
	}
}
