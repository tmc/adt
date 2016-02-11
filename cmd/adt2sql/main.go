// Command
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tmc/adt"
)

var (
	flagFile      = flag.String("f", "", "path to ADT file")
	flagTableName = flag.String("n", "", "name of resulting database table")
	flagVerbose   = flag.Bool("v", false, "verbose")
)

func main() {
	flag.Parse()
	if err := migrate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("No value present in DATABASE_URL environment variable.")
	}
	db, err := newDBFromURL(dbURL)
	if err != nil {
		return err
	}

	table, err := adt.TableFromPath(*flagFile)
	if err != nil {
		return err
	}

	if *flagTableName == "" {
		*flagTableName = strings.TrimSuffix(*flagFile, ".ADT")
	}

	ddl, err := table.SQLDDL(*flagTableName)
	if err != nil {
		return err
	}

	if *flagVerbose {
		fmt.Println(ddl)
	}
	_, err = db.Exec(ddl)
	if err != nil {
		return err
	}

	prepped, err := db.Preparex(table.InsertSQL(*flagTableName))
	if err != nil {
		return err
	}

	for i := 0; i < int(table.RecordCount); i++ {
		r, err := table.Get(i)
		if err != nil {
			return err
		}

		values := make([]interface{}, 0, len(table.Columns))
		for _, column := range table.Columns {
			values = append(values, r[column.Name])
		}
		if _, err = prepped.Exec(values...); err != nil {
			return err
		}
	}
	log.Println(table.RecordCount, "rows inserted")

	return nil
}

func newDBFromURL(URL string) (*sqlx.DB, error) {
	p, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	DSN := strings.TrimLeft(p.String(), p.Scheme+"://")
	return sqlx.Connect(p.Scheme, DSN)
}
