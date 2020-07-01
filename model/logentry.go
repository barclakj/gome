package model

import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	Data           []byte
	Origin         string
	Seq            uint64
	Oid            string
	Hash           string
	Ts             int64
	OriginTs       int64
	Branch         int64
	PreviousBranch int64
}

type LogEntryObservers struct {
	Oid       string
	Observers []string
}

type LogEntryCommand struct {
	Oid     string
	Branch  int64
	Command string
	Origin  string
	Hash    string
}

const OBSERVE_COMMAND = `observe` // notifies that we're interested in observing the entity.
const REPLAY_COMMAND = `replay`   // notify replay all events for the object/branch
const IGNORE_COMMAND = `ignore`   // notifies that we want to stop getting events fo the entity
const SYNC_COMMAND = `sync`       // broadcasts the latest hash for the branch for comparison

const CMD_OID = `urn:uuid:1`

type bySeqAndTs []LogEntry

func (a bySeqAndTs) Len() int      { return len(a) }
func (a bySeqAndTs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a bySeqAndTs) Less(i, j int) bool {
	if a[i].Seq < a[j].Seq {
		return true
	} else if a[i].Seq == a[j].Seq {
		if a[i].Ts < a[j].Ts {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

/* Creates a new log entry - complete with defined origin and data */
func NewLogEntry(data []byte, origin string) *LogEntry {
	le := LogEntry{}

	id := uuid.New()

	le.Oid = id.URN()
	le.Seq = 1
	le.Ts = time.Now().UnixNano()
	le.OriginTs = le.Ts
	le.Origin = origin
	le.Data = data
	le.Branch = 0
	le.PreviousBranch = -1

	log.Printf("Oid %s\n", le.Oid)

	return &le
}

func (le *LogEntry) Validate() bool {
	if len(le.Oid) > 0 && le.Seq >= 1 && le.Branch >= 0 && le.Data != nil && len(le.Data) > 0 && len(le.Origin) > 0 {
		return true
	} else {
		return false
	}
}

/* Converts a log entry to a string json representing. */
func (le *LogEntry) ToJSON() string {
	jsonMap, _ := json.Marshal(le)
	return string(jsonMap)
}

/* Converts a json string of a log-entry to a log entry. */
func FromJSON(jsonString []byte) *LogEntry {
	le := LogEntry{}
	json.Unmarshal(jsonString, &le)
	return &le
}

/* Sorts an array of log entries by sequence and timestamp */
func Sort(logs []LogEntry) {
	sort.Sort(bySeqAndTs(logs))
}

/* Converts a log entry command to a string json representing. */
func (cmd *LogEntryCommand) ToJSON() string {
	jsonMap, _ := json.Marshal(cmd)
	return string(jsonMap)
}

/* Converts a json string of a log-entry-command to a log entry command. */
func CmdFromJSON(jsonString []byte) *LogEntryCommand {
	cmd := LogEntryCommand{}
	json.Unmarshal(jsonString, &cmd)
	return &cmd
}
