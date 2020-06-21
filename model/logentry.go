package model

import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	Data     []byte
	Origin   string
	Seq      uint64
	Oid      string
	Hash     string
	Ts       int64
	OriginTs int64
}

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
	if data == nil || origin == "" {
		return nil
	} else {
		le := LogEntry{}

		id := uuid.New()

		le.Oid = id.URN()
		le.Seq = 1
		le.Ts = time.Now().UnixNano()
		le.OriginTs = le.Ts
		le.Origin = origin
		le.Data = data

		log.Printf("Oid %s\n", le.Oid)

		return &le
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
