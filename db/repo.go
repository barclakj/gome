package db

import  (
	"realizr.io/gome/model"
	"log"
	// "github.com/google/uuid"
	"database/sql"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

const DB_FILENAME = "gome.db"

const INSERT_LE_SQL = `INSERT INTO GOME_LOG ("origin", "uuid", "seq", "data", "remote_ts", "ts") VALUES (?, ?, ?, ?, ?, ?);`

const CREATE_LE_SQL = `CREATE TABLE GOME_LOG (
		"origin" VARCHAR(100) NOT NULL,
		"uuid" VARCHAR(100) NOT NULL,
		"seq" INTEGER NOT NULL,
		"data" BLOB,
		"remote_ts" LONG NOT NULL,
		"ts" LONG NOT NULL);`

const CREATE_LE_PK = `CREATE UNIQUE INDEX log_pk ON GOME_LOG("origin", "uuid", "seq", "remote_ts");`

const QUERY_LE_BY_UUID = `SELECT "origin", "uuid", "seq", "data", "remote_ts", "ts" FROM GOME_LOG WHERE "uuid" = ? ORDER BY "ts" ASC;`

func createDB() {
	file, err := os.Create(DB_FILENAME)
	if err!=nil {
		log.Fatal(err.Error())
	}
	file.Close()

	db, err  := sql.Open("sqlite3", DB_FILENAME)
	if err!=nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	statement, err :=  db.Prepare(CREATE_LE_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
}


func openDB() {
	if _, err := os.Stat(DB_FILENAME); os.IsNotExist(err) {
		createDB()
	}
	db, _ := sql.Open("sqlite3", DB_FILENAME)

	database = db
}

func insertLogEntry(le *model.LogEntry) bool {
	statement, err := database.Prepare(INSERT_LE_SQL)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec(le.Origin, le.Uuid, le.Seq, le.Data, le.RemoteTs, le.Ts)
	if err != nil {
		log.Fatal(err.Error())
		return false
	} else {
		return true
	}
}

func Load(ref  string) []model.LogEntry {
	row, err := database.Query(QUERY_LE_BY_UUID, ref)
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	var logs []model.LogEntry

	for row.Next() {
		log.Printf("Scanning next record\n")
		le := model.LogEntry{}
		row.Scan(&le.Origin, &le.Uuid, &le.Seq, &le.Data, &le.RemoteTs, &le.Ts)
		logs = append(logs, le)
	}
	model.Sort(logs)
	return logs
}

func Append(le *model.LogEntry) bool {
	if database == nil {
		openDB()
	}
	if le.Origin == "" || le.Uuid == "" || le.Seq <= 0 || le.RemoteTs <= 0 {
		log.Printf("Log entry is not valid!\n")
		return false
	} else {
		gl := Load(le.Uuid)
		for _, tle := range gl {
			log.Printf("Testing %s %d\n", tle.Origin, tle.Seq)
			if tle.Origin == le.Origin && tle.Seq == le.Seq {
				log.Fatal("Existing entry found\n")
				return false
			}
		}
		log.Printf("Appending %s:%d from %s\n", le.Uuid, le.Seq, le.Origin)
		return insertLogEntry(le)
	}
}
