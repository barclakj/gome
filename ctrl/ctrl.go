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
const MAX_WAIT_SECONDS = 10

type LogEntryController struct {
	Alive bool
}

/* Initializer/constructor */
func (ctrl *LogEntryController) Init(wg *sync.WaitGroup) {
	ctrl.Alive = true
	go ctrl.listen(wg)
	time.Sleep(2 * time.Second)
}

/* Notifies all subscribers of a change to an object. */
func (ctrl *LogEntryController) notifyInterestedParties(le *model.LogEntry) {
	observers := []string{"192.168.86.255:7456"}

	registeredObservers := db.LoadAllObservers(le.Oid)
	observers = append(observers, registeredObservers.Observers...)
	log.Printf("Raising notification for %d observers on object %s\n", len(observers), le.Oid)

	broadcast.Send(le, observers)
}

/* Inserts a new log entry. Note that a log entry is immutable so this appends a new version to the list of entries. */
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

/* Saves a new log entry with provided data. This is for use locally only (within the server) */
func (ctrl *LogEntryController) Save(data []byte) *model.LogEntry {
	origin := env.GetOrigin()
	le := model.NewLogEntry(data, origin)
	le.Hash = model.Hash(le.Oid, "", le.Data)
	le = ctrl.insert(le)
	return le
}

/* Updates (appends) to an existing log entry with provided data. This is for use locally only (within the server) */
func (ctrl *LogEntryController) Update(oid string, branch uint64, data []byte) *model.LogEntry {
	currentLogEntry := db.FetchLatestLogEntry(oid, branch)
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

/* Handles a message from a remote server.
To be valid the object must either be new or the hash must match expectations.
Those that do not match will be rejected.
TODO: Change this so rejected entries are stored along a separate timeline so that someone
could merge these in later if desired.
*/
func (ctrl *LogEntryController) fromRemote(le *model.LogEntry, remoteAddress string) *model.LogEntry {
	branch := uint64(0)
	currentLogEntries := db.FetchLogEntries(le.Oid, le.Seq-1)

	for _, currentLogEntry := range currentLogEntries {
		newHash := model.Hash(currentLogEntry.Oid, currentLogEntry.Hash, le.Data)
		if le.Hash == newHash { // we've found a branch
			le.Ts = time.Now().UnixNano()
			le.Branch = currentLogEntry.Branch
			log.Printf("Updating %s from %s!\n", le.Oid, remoteAddress)
			return ctrl.insert(le)
		}
		branch = currentLogEntry.Branch + 1
	}
	le.Ts = time.Now().UnixNano()
	le.Branch = branch
	log.Printf("Creating %s from %s!\n", le.Oid, remoteAddress)
	return ctrl.insert(le)

}

/* Listens for new messages from other servers. Only messages which are not from the local host are accepted. */
func (ctrl *LogEntryController) listen(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	addr, _ := net.ResolveUDPAddr("udp", CMD_PORT)
	sock, _ := net.ListenUDP("udp", addr)
	if sock != nil {
		defer sock.Close()
		log.Printf("Listening on %s\n", addr)

		i := 0
		buf := make([]byte, MAX_MSG_SIZE)
		for ctrl.Alive {
			sock.SetReadDeadline(time.Now().Add(time.Second * MAX_WAIT_SECONDS))
			rlen, remoAddr, err := sock.ReadFromUDP(buf)
			i += rlen
			if err != nil {
				fmt.Println(err)
			} else {
				if !env.IsLocalAddress(remoAddr.IP.String()) {
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
