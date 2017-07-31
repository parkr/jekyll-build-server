package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	mySQLDateTimeFormat = "2006-01-02 15:04:05"
	schema              = `CREATE TABLE builds (
	    id varchar(255) NOT NULL DEFAULT '',
	    output text NOT NULL,
	    success tinyint(1) NOT NULL DEFAULT '0',
	    created_at datetime NOT NULL,
	    completed_at datetime NOT NULL,
	    PRIMARY KEY (id),
	    UNIQUE KEY id_uniq (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`
	checkIfSchemaExists = `SELECT COUNT(*) as does_exist FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'builds';`

	selectBuild   = `SELECT * FROM builds WHERE id=?;`
	insertBuild   = `INSERT INTO builds (id, output, success, created_at, completed_at) VALUES (:id, :output, :success, :created_at, :completed_at);`
	updateOutput  = `UPDATE builds SET output = CONCAT_WS(CHAR(10 using utf8), output, :line) WHERE id = :id;`
	updateBuild   = `UPDATE builds SET output=:output, success=:success, created_at=:created_at, completed_at=:completed_at WHERE id=:id;`
	completeBuild = `UPDATE builds SET completed_at = :completed_at WHERE id = :id;`
)

var (
	db *sqlx.DB

	dbConnString string
)

func mySQLFormattedTime() string {
	return time.Now().UTC().Format(mySQLDateTimeFormat)
}

func init() {
	flag.StringVar(&dbConnString, "db", "", "Connection string for database. Leave blank to omit db logging.")
}

type TableCheck struct {
	DoesExist int `db:"does_exist"`
}

func InitDatabase() {
	db = sqlx.MustConnect("mysql", dbConnString)
	db.Ping()
	var check TableCheck
	err := db.Get(&check, checkIfSchemaExists)
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}
	if check.DoesExist < 1 {
		db.MustExec(schema)
	}
}

type OutputUpdate struct {
	Id   string `db:"id"`
	Line string `db:"line"`
}

type Build struct {
	Id          string `db:"id"`
	Output      string `db:"output"`
	Success     bool   `db:"success"`
	CreatedAt   string `db:"created_at"`
	CompletedAt string `db:"completed_at"`
	saved       bool
}

func (b *Build) Exists() bool {
	return b.Get(b.Id) != nil || b.saved
}

func (b *Build) Get(id string) error {
	if db == nil {
		return nil
	}

	if id != "" {
		b.Id = id
	}

	err := db.Get(b, selectBuild, id)
	if err != nil {
		log.Printf("[%s] db: Get() received an error: %v", b.Id, err)
		b.saved = false
		return err
	}
	b.saved = true
	return nil
}

func (b *Build) UpdateOutput(line string) error {
	if db == nil {
		return nil
	}

	if b.saved {
		_, err := db.NamedExec(updateOutput, &OutputUpdate{
			Id:   b.Id,
			Line: line,
		})
		return err
	} else {
		return b.Save()
	}
}

func (b *Build) Log(msg string) {
	msg = fmt.Sprintf("%s %s", mySQLFormattedTime(), msg)
	if b.Output == "" {
		b.Output = msg
	} else {
		b.Output = b.Output + "\n" + msg
	}
	if err := b.UpdateOutput(msg); err != nil {
		log.Printf("[%s] db: error saving log message: %v", b.Id, err)
	}
}

func (b *Build) Save() error {
	if db == nil {
		return nil
	}

	if b.CompletedAt == "" {
		b.CompletedAt = time.Unix(0, 0).UTC().Format(mySQLDateTimeFormat)
		log.Printf("[%s]: db: new completed_at value: %s", b.Id, b.CompletedAt)
	}

	log.Printf("[%s] db: has been saved: %v", b.Id, b.saved)

	var query string
	if b.saved {
		log.Printf("[%s] db: updating the build", b.Id)
		query = updateBuild
	} else {
		log.Printf("[%s] db: inserting the build", b.Id)
		query = insertBuild
	}

	_, err := db.NamedExec(query, b)
	if err == nil {
		b.saved = true
	}
	return err
}
