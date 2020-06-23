package db

import (
	"log"

	"realizr.io/gome/env"

	"realizr.io/gome/model"

	// "github.com/google/uuid"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

const DB_FILENAME = "/gome.db"

const INSERT_LE_SQL = `INSERT INTO GOME_LOG ("origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch") VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

const CREATE_LE_SQL = `CREATE TABLE GOME_LOG (
	"oid" VARCHAR(100) NOT NULL,
	"seq" INTEGER NOT NULL,
	"origin" VARCHAR(100) NOT NULL,
	"origin_ts" LONG NOT NULL,
	"data" BLOB,
	"hash" VARCHAR(256),
	"branch" LONG NOT NULL,
	"ts" LONG NOT NULL);`

const CREATE_LE_OBSERVER_SQL = `CREATE TABLE GOME_LOG_OBSERVER (
	"oid" VARCHAR(100) NOT NULL,
	"observer" VARCHAR(378) NOT NULL);`

const CREATE_LE_PK = `CREATE UNIQUE INDEX log_pk ON GOME_LOG("oid", "seq", "branch");`

const CREATE_LE_OBSERVER_PK = `CREATE UNIQUE INDEX log_obs_pk ON GOME_LOG_OBSERVER("oid", "observer");`

const QUERY_LE_BY_OID = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch" FROM GOME_LOG WHERE "oid" = ? AND "branch" = ? ORDER BY "seq" ASC;`

const QUERY_LE_BY_OID_AND_SEQ = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch" FROM GOME_LOG WHERE "oid" = ? AND "seq" = ? ORDER BY "batch" ASC;`

const QUERY_LE_BY_LATEST_OID = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch" FROM GOME_LOG WHERE "oid" = ? AND "branch" = ? ORDER BY "seq" DESC LIMIT 1;`

const QUERY_LE_OBSERVERS = `SELECT "observer" FROM GOME_LOG_OBSERVER WHERE "oid" = ?;`

const INSERT_LE_OBS_SQL = `INSERT INTO GOME_LOG_OBSERVER ("oid", "observer") VALUES (?, ?);`

const DELETE_LE_OBS_SQL = `DELETE FROM GOME_LOG_OBSERVER WHERE "oid"=? AND "observer"=?;`

/* Creates a new DB */
func createDB() {
	filename := env.GetHome() + DB_FILENAME
	log.Printf(filename)
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	statement, err := db.Prepare(CREATE_LE_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()

	statement, err = db.Prepare(CREATE_LE_PK)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()

	statement, err = db.Prepare(CREATE_LE_OBSERVER_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()

	statement, err = db.Prepare(CREATE_LE_OBSERVER_PK)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
}

func openDB() {
	filename := env.GetHome() + DB_FILENAME
	if database == nil {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			createDB()
		}
		db, _ := sql.Open("sqlite3", filename)

		database = db
	}
}

func InsertLogEntry(le *model.LogEntry) bool {
	openDB()
	statement, err := database.Prepare(INSERT_LE_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(le.Origin, le.Oid, le.Seq, le.Data, le.Hash, le.OriginTs, le.Ts, le.Branch)
	if err != nil {
		log.Fatal(err.Error())
		return false
	} else {
		return true
	}
}

func FetchLogEntries(ref string, seq uint64) []model.LogEntry {
	openDB()
	row, err := database.Query(QUERY_LE_BY_OID_AND_SEQ, ref, seq)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var logEntries []model.LogEntry

	for row.Next() {
		le := model.LogEntry{}
		log.Printf("Scanning next record ")
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch)
		log.Printf("...found %s\n", le.Oid)

		logEntries = append(logEntries, le)
	}
	return logEntries
}

func FetchLatestLogEntry(ref string, branch uint64) *model.LogEntry {
	openDB()
	row, err := database.Query(QUERY_LE_BY_LATEST_OID, ref, branch)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	le := model.LogEntry{}

	for row.Next() {
		log.Printf("Scanning next record ")
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch)
		log.Printf("...found %s\n", le.Oid)
		break
	}
	if le.Oid == "" {
		var none *model.LogEntry
		return none
	} else {
		return &le
	}
}

func LoadAllLogEntries(ref string, branch uint64) []model.LogEntry {
	openDB()
	row, err := database.Query(QUERY_LE_BY_OID, ref, branch)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var logs []model.LogEntry

	for row.Next() {
		log.Printf("Scanning next record\n")
		le := model.LogEntry{}
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch)
		logs = append(logs, le)
	}
	model.Sort(logs)
	return logs
}

func LoadAllObservers(ref string) model.LogEntryObservers {
	openDB()
	row, err := database.Query(QUERY_LE_OBSERVERS, ref)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	observers := model.LogEntryObservers{}
	observers.Oid = ref

	for row.Next() {
		log.Printf("Observer...\n")
		var obs string
		row.Scan(&obs)
		observers.Observers = append(observers.Observers, obs)
	}
	return observers
}

func AddObserver(ref string, observer string) bool {
	openDB()
	statement, err := database.Prepare(INSERT_LE_OBS_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(ref, observer)
	if err != nil {
		log.Printf(err.Error())
		return false
	} else {
		return true
	}
}

func RemoveObserver(ref string, observer string) bool {
	openDB()
	statement, err := database.Prepare(DELETE_LE_OBS_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(ref, observer)
	if err != nil {
		log.Printf(err.Error())
		return false
	} else {
		return true
	}
}
