package broadcast

import (
	"realizr.io/gome/model"
	"log"
	"net"
)

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
	log.Printf("tx=>%s\n", address);
}
