package ctrl

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"realizr.io/gome/broadcast"

	"realizr.io/gome/env"

	"realizr.io/gome/db"
	"realizr.io/gome/model"
)

const MAX_MSG_SIZE = 2048000
const CMD_PORT = "0.0.0.0:7456"
const MAX_WAIT_SECONDS = 1

type LogEntryController struct {
	Alive bool
}

func (ctrl *LogEntryController) Init(wg *sync.WaitGroup) {
	ctrl.Alive = true
	go ctrl.Listen(wg)
	time.Sleep(2 * time.Second)
}

func (ctrl *LogEntryController) IsAlive() bool {
	return ctrl.Alive
}

func (ctrl *LogEntryController) notifyInterestedParties(le *model.LogEntry) {
	subscribers := []string{"192.168.86.255:7456"}
	broadcast.Send(le, subscribers)
}

func (ctrl *LogEntryController) insert(le *model.LogEntry) *model.LogEntry {
	success := db.InsertLogEntry(le)
	if success {
		go ctrl.notifyInterestedParties(le)
		return le
	} else {
		log.Printf("Failed to save log entry...\n")
		var none *model.LogEntry
		return none
	}
}

/* Saves a new log entry with provided data */
func (ctrl *LogEntryController) Save(data []byte) *model.LogEntry {
	origin := env.GetOrigin()
	le := model.NewLogEntry(data, origin)
	le.Hash = model.Hash(le.Oid, "", le.Data)
	le = ctrl.insert(le)
	return le
}

/* Saves a new log entry with provided data */
func (ctrl *LogEntryController) Update(oid string, data []byte) *model.LogEntry {
	currentLogEntry := db.FetchLatestLogEntry(oid)
	if currentLogEntry != nil {
		currentLogEntry.Origin = env.GetOrigin()
		newHash := model.Hash(currentLogEntry.Oid, currentLogEntry.Hash, data)
		currentLogEntry.Hash = newHash
		currentLogEntry.Seq = currentLogEntry.Seq + 1
		currentLogEntry.Ts = time.Now().UnixNano()
		currentLogEntry.OriginTs = time.Now().UnixNano()
		currentLogEntry = ctrl.insert(currentLogEntry)
	} else {
		currentLogEntry = ctrl.Save(data)
	}
	return currentLogEntry
}

func (ctrl *LogEntryController) fromRemote(le *model.LogEntry, remoteAddress string) *model.LogEntry {
	currentLogEntry := db.FetchLatestLogEntry(le.Oid)
	if currentLogEntry != nil {
		newHash := model.Hash(currentLogEntry.Oid, currentLogEntry.Hash, le.Data)
		if le.Hash == newHash {
			le.Ts = time.Now().UnixNano()
			log.Printf("Updating %s from %s!\n", le.Oid, remoteAddress)
			currentLogEntry = ctrl.insert(le)
		} else {
			log.Printf("Rejecting update on %s from %s due to hash mismatch!\n", le.Oid, remoteAddress)
			currentLogEntry = nil
		}
	} else {
		le.Ts = time.Now().UnixNano()
		log.Printf("Creating %s from %s!\n", le.Oid, remoteAddress)
		currentLogEntry = ctrl.insert(le)
	}
	return currentLogEntry

}

func (ctrl *LogEntryController) Listen(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	addr, _ := net.ResolveUDPAddr("udp", CMD_PORT)
	sock, _ := net.ListenUDP("udp", addr)
	if sock != nil {
		defer sock.Close()
		log.Printf("Listening on %s\n", addr)

		i := 0
		buf := make([]byte, MAX_MSG_SIZE)
		for ctrl.IsAlive() {
			sock.SetReadDeadline(time.Now().Add(time.Second * MAX_WAIT_SECONDS))
			rlen, remoAddr, err := sock.ReadFromUDP(buf)
			i += rlen
			if err != nil {
				fmt.Println(err)
			} else {
				if !isLocalAddress(remoAddr.IP.String()) {
					le := model.FromJSON(buf[0:rlen])
					log.Printf("JSON %s\n", le.ToJSON())
					ctrl.fromRemote(le, remoAddr.IP.String())
				}
			}
		}
	} else {
		log.Printf("Cannot listen on port %s. Probably in use!\n", CMD_PORT)
	}
}

func isLocalAddress(testIP string) bool {
	local := false
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.String() == testIP {
				log.Printf("Local msg rcvd. Ignoring...")
				local = true
				break
			}
			// process IP address
		}
		if local == true {
			break
		}
	}
	return local
}
