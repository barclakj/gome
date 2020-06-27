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
	var rval *model.LogEntry
	if le.Validate() {
		success := db.InsertLogEntry(le)
		if success {
			go ctrl.notifyInterestedParties(le)
			rval = le
		} else {
			log.Printf("Failed to save log entry...\n")
		}
	} else {
		log.Printf("Log entry is not valid.. refusing to save!!!")
	}
	return rval
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

func (ctrl *LogEntryController) processStashedLogEntries(oid string, seq uint64) *model.LogEntry {
	var ultLe *model.LogEntry
	log.Printf("Checking for any stashed log entries...")
	stashedMap := db.FetchStashedLogEntries(oid, seq+1)
	for sid, le := range stashedMap {
		ultLe = ctrl.FromRemote(le)
		if ultLe != nil {
			db.DeleteStash(sid)
		}
	}
	return ultLe
}

/* Handles a message from a remote server.
To be valid the object must either be new or the hash must match expectations.
Those that do not match will be rejected.
TODO: Change this so rejected entries are stored along a separate timeline so that someone
could merge these in later if desired.
*/
func (ctrl *LogEntryController) FromRemote(le *model.LogEntry) *model.LogEntry {
	done := false
	var rval *model.LogEntry

	if le.Validate() {
		if db.CheckLogEntryExistsByHash(le.Oid, le.Seq, le.Hash) {
			log.Printf("Ignoring duplicate...")
		} else if le.Seq == 1 {
			// why do you think I'm interested in this?
			if !db.CheckLogEntryExistsByBranch(le.Oid, le.Seq, 0) {
				// doesn't exist.. ok, well lets add it then if the hash looks valid..
				if le.Hash == model.Hash(le.Oid, "", le.Data) {
					// ok...
					log.Printf("Allowing new object from %s as %s", le.Origin, le.Oid)
					rval = ctrl.insert(le)
				}
			} else {
				// exists so this is like a new instance of a log but we have the uuid already!? nuh-uh..
				// this would be like someones spamming a conflicting object or we've got a uuid duplicate.
				log.Printf("%s tried to create a duplicate log for %s!? REJECTING!", le.Origin, le.Oid)
			}
		} else {
			currentLogEntries := db.FetchLogEntries(le.Oid, le.Seq-1) // find those which this one should be based on

			for _, currentLogEntry := range currentLogEntries {
				newHash := model.Hash(currentLogEntry.Oid, currentLogEntry.Hash, le.Data)
				log.Printf("Testing hash: %s = %s ?", le.Hash, newHash)
				if le.Hash == newHash { // we've found a branch
					if db.CheckLogEntryExistsByBranch(le.Oid, le.Seq, le.Branch) {
						// already exists.. need to branch.
						le.Branch = db.GetNextBranch(le.Oid)
						le.PreviousBranch = currentLogEntry.Branch
						log.Printf("Branching %s from %s!\n", le.Oid, le.Origin)
					} else {
						// can append to log
						le.Branch = currentLogEntry.Branch
						le.PreviousBranch = currentLogEntry.PreviousBranch
						log.Printf("Appending %s from %s!\n", le.Oid, le.Origin)
					}
					le.Ts = time.Now().UnixNano()
					rval = ctrl.insert(le)
					done = true
					break
				}
			}
			// if I've not saved it then it's not based on anything I know about...
			// this could be because its based on another branch from something else..
			// stash the message as out of order and reprocess later...
			if !done {
				log.Printf("Stashing potential out of order message...(not yet)")
				db.StashLogEntry(le)
			} else {
				rval = ctrl.processStashedLogEntries(le.Oid, le.Seq)
				if rval == nil {
					rval = le
				}
			}
		}
	} else {
		log.Printf("Log is not valid! Rejecting...")
	}

	return rval

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
					le.Origin = remoAddr.IP.String()
					log.Printf("JSON %s\n", le.ToJSON())
					go ctrl.FromRemote(le)
				}
			}
		}
	} else {
		log.Printf("Cannot listen on port %s. Probably in use!\n", CMD_PORT)
	}
}
