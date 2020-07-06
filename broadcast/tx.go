package broadcast

import (
	"log"
	"net"
	"sync"

	"realizr.io/gome/model"
)

func Send(logEntry *model.LogEntry, subscribers []string, wg *sync.WaitGroup) {
	dat := []byte(logEntry.ToJSON())
	log.Printf("Sending %s", string(dat))

	for _, element := range subscribers {
		go tx(element, dat, wg)
	}
}

func tx(address string, msg []byte, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	log.Printf("tx=>%s\n", address)
	conn, _ := net.Dial("udp", address)
	defer conn.Close()
	conn.Write(msg)
}
