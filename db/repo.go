package db

import (
	"log"

	"realizr.io/gome/env"

	"realizr.io/gome/model"

	// "github.com/google/uuid"
	"database/sql"
	"os"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

const DB_FILENAME = "/gome.db"

const INSERT_LE_SQL = `INSERT INTO GOME_LOG ("origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch", "previous_branch") VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

const CREATE_LE_SQL = `CREATE TABLE GOME_LOG (
	"oid" VARCHAR(100) NOT NULL,
	"seq" INTEGER NOT NULL,
	"origin" VARCHAR(100) NOT NULL,
	"origin_ts" LONG NOT NULL,
	"data" BLOB,
	"hash" VARCHAR(256),
	"branch" LONG NOT NULL,
	"previous_branch" LONG,
	"ts" LONG NOT NULL);`

const CREATE_LE_OBSERVER_SQL = `CREATE TABLE GOME_LOG_OBSERVER (
	"oid" VARCHAR(100) NOT NULL,
	"observer" VARCHAR(378) NOT NULL);`

const CREATE_LE_STASH_SQL = `CREATE TABLE GOME_STASH (
	"sid" VARCHAR(100) NOT NULL PRIMARY KEY,
	"oid" VARCHAR(100) NOT NULL,
	"seq" INTEGER NOT NULL,
	"msg" BLOB);`

const CREATE_LE_STASH_IDX = `CREATE INDEX log_stash_idx ON GOME_STASH("oid", "seq");`

const CREATE_LE_PK = `CREATE UNIQUE INDEX log_pk ON GOME_LOG("oid", "seq", "branch");`

const CREATE_LE_OBSERVER_PK = `CREATE UNIQUE INDEX log_obs_pk ON GOME_LOG_OBSERVER("oid", "observer");`

const QUERY_LE_BY_OID = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch", "previous_branch" FROM GOME_LOG WHERE "oid" = ? AND "branch" = ? ORDER BY "seq" ASC;`

const QUERY_LE_BY_OID_AND_SEQ = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch", "previous_branch" FROM GOME_LOG WHERE "oid" = ? AND "seq" = ? ORDER BY "batch" ASC;`

const QUERY_LE_BY_LATEST_OID = `SELECT "origin", "oid", "seq", "data", "hash", "origin_ts", "ts", "branch", "previous_branch" FROM GOME_LOG WHERE "oid" = ? AND "branch" = ? ORDER BY "seq" DESC LIMIT 1;`

const QUERY_LE_OBSERVERS = `SELECT "observer" FROM GOME_LOG_OBSERVER WHERE "oid" = ?;`

const INSERT_LE_OBS_SQL = `INSERT INTO GOME_LOG_OBSERVER ("oid", "observer") VALUES (?, ?);`

const DELETE_LE_OBS_SQL = `DELETE FROM GOME_LOG_OBSERVER WHERE "oid"=? AND "observer"=?;`

const QUERY_EXISTS_BY_BRANCH = `SELECT 1 FROM GOME_LOG WHERE "oid" = ? AND "seq" = ? AND "branch" = ?;`

const QUERY_EXISTS_BY_HASH = `SELECT 1 FROM GOME_LOG WHERE "oid" = ? AND "seq" = ? AND "hash" = ?;`

const QUERY_MAX_BRANCH = `SELECT max(branch) FROM GOME_LOG WHERE "oid" = ?;`

const QUERY_LE_STASH_SQL = `SELECT sid, msg FROM GOME_STASH WHERE "oid" = ? AND "seq" = ?`

const DELETE_LE_STASH_SQL = `DELETE FROM GOME_STASH WHERE "sid"=?;`

const INSERT_LE_STASH_SQL = `INSERT INTO GOME_STASH ("sid", "oid", "seq", "msg") VALUES (?, ?, ?, ?);`

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

	statement, err = db.Prepare(CREATE_LE_STASH_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()

	statement, err = db.Prepare(CREATE_LE_STASH_IDX)
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
	_, err = statement.Exec(le.Origin, le.Oid, le.Seq, le.Data, le.Hash, le.OriginTs, le.Ts, le.Branch, le.PreviousBranch)
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
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch, &le.PreviousBranch)
		log.Printf("...found %s\n", le.Oid)

		logEntries = append(logEntries, le)
	}
	return logEntries
}

func FetchLatestLogEntry(ref string, branch int64) *model.LogEntry {
	openDB()
	row, err := database.Query(QUERY_LE_BY_LATEST_OID, ref, branch)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	le := model.LogEntry{}

	for row.Next() {
		log.Printf("Scanning next record ")
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch, &le.PreviousBranch)
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

func CheckLogEntryExistsByBranch(ref string, seq uint64, branch int64) bool {
	openDB()
	exists := false
	row, err := database.Query(QUERY_EXISTS_BY_BRANCH, ref, seq, branch)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		log.Printf("Record exists..")
		exists = true
		break // no need to continue
	}
	return exists
}

func CheckLogEntryExistsByHash(ref string, seq uint64, hash string) bool {
	openDB()
	exists := false
	row, err := database.Query(QUERY_EXISTS_BY_HASH, ref, seq, hash)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		log.Printf("Record exists..")
		exists = true
		break // no need to continue
	}
	return exists
}

func GetNextBranch(ref string) int64 {
	openDB()
	maxBranch := int64(0)
	row, err := database.Query(QUERY_MAX_BRANCH, ref)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		row.Scan(&maxBranch)
		maxBranch++
		break // no need to continue
	}
	log.Printf("Max branch for %s = %d", ref, maxBranch)

	return maxBranch
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
		row.Scan(&le.Origin, &le.Oid, &le.Seq, &le.Data, &le.Hash, &le.OriginTs, &le.Ts, &le.Branch, &le.PreviousBranch)
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

func StashLogEntry(le *model.LogEntry) bool {
	openDB()
	statement, err := database.Prepare(INSERT_LE_STASH_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	sid := uuid.New()
	leJSON := le.ToJSON()
	log.Printf("Stashing sid %s for entry %s, %d", sid, le.Oid, le.Seq)
	_, err = statement.Exec(sid, le.Oid, le.Seq, leJSON)
	if err != nil {
		log.Fatal(err.Error())
		return false
	} else {
		return true
	}
}

func FetchStashedLogEntries(ref string, seq uint64) map[string]*model.LogEntry {
	log.Printf("Fetching stashed entries for %s, %d", ref, seq)
	openDB()
	row, err := database.Query(QUERY_LE_STASH_SQL, ref, seq)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var m map[string]*model.LogEntry
	m = make(map[string]*model.LogEntry)
	var json string
	var sid string

	for row.Next() {
		log.Printf("Scanning next record\n")
		row.Scan(&sid, &json)
		le := model.FromJSON([]byte(json))
		m[sid] = le
	}
	log.Printf("Found %d stashed entries.", len(m))
	return m
}

func DeleteStash(sid string) bool {
	log.Printf("Deleting stashed %s", sid)
	openDB()
	statement, err := database.Prepare(DELETE_LE_STASH_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(sid)
	if err != nil {
		log.Fatal(err.Error())
		return false
	} else {
		return true
	}
}
