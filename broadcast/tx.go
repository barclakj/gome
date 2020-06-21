package broadcast

import (
	"log"
	"net"

	"realizr.io/gome/model"
)

func Send(logEntry *model.LogEntry, subscribers []string) {
	dat := []byte(logEntry.ToJSON())

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
