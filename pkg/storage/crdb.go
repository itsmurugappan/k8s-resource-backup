package storage

import (
	"log"
	"time"

	"database/sql"
	_ "github.com/lib/pq"
)

type CRDB struct {
	ConnectionString string
}

func (db *CRDB) Store(data []byte, ns string, snap_id string, name string) {

	conn, err := sql.Open("postgres", db.ConnectionString)

	if err != nil {
		log.Printf("error connecting to the database: %s", err)
	}
	defer conn.Close()

	// delete existing snaphots if any
	conn.Exec("Delete from resource_backup where snap_id=$1 and ns=$2 and name=$3", snap_id, ns, name)

	if _, err := conn.Exec(
		"INSERT INTO resource_backup (ns, snap_id, resource, name, updatedtime) VALUES ($1, $2, $3, $4, $5)", ns, snap_id, string(data), name, time.Now()); err != nil {
		log.Printf("error inserting into to the database: %s", err)
	}
}

func (db *CRDB) Retrieve(ns string, snap_id string, name string) []byte {

	conn, err := sql.Open("postgres", db.ConnectionString)

	if err != nil {
		log.Printf("error connecting to the database: %s", err)
	}
	defer conn.Close()

	rows, err := conn.Query("select resource from resource_backup where snap_id=$1 and ns=$2 and name=$3", snap_id, ns, name)
	if err != nil {
		log.Printf("error getting snapshot from db : %s", err)
	}
	defer rows.Close()

	var result []byte
	rows.Next() // take the first row
	if err := rows.Scan(&result); err != nil {
		log.Print("record not found")
		return []byte("")
	}
	return result
}
