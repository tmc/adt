// Command
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jmoiron/sqlx"
	"github.com/tmc/adt"
)

var (
	flagFile       = flag.String("f", "", "path to ADT file")
	flagTableName  = flag.String("n", "", "name of resulting database table")
	flagVerbose    = flag.Bool("v", false, "verbose")
	flagMinRecords = flag.Int("minrecords", 1, "if a table has fewer than this many records it will be skipped")
)

func main() {
	flag.Parse()
	if err := migrate(); err != nil {
		log.Println(*flagTableName, err)
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
	if *flagTableName == "" {
		*flagTableName = strings.TrimSuffix(*flagFile, ".ADT")
	}

	table, err := adt.TableFromPath(*flagFile)
	if err != nil {
		return err
	}
	if int(table.RecordCount) < *flagMinRecords {
		return fmt.Errorf("too few records (%d)", table.RecordCount)
	}

	ddl, err := table.SQLDDL(*flagTableName)
	if err != nil {
		return err
	}

	if *flagVerbose {
		fmt.Println(ddl)
	}
	if _, err = db.Exec(ddl); err != nil {
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
			var value interface{} = r[column.Name]
			if !reflect.ValueOf(value).IsValid() {
				value = nil
			}
			if dur, ok := value.(time.Duration); ok {
				value = durToSQL(dur)
			}
			values = append(values, value)
		}
		if _, err = prepped.Exec(values...); err != nil {
			fmt.Println("insert error")
			spew.Dump(r)
			return err
		}
	}
	log.Println(*flagTableName, table.RecordCount, "rows inserted")

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

func durToSQL(d time.Duration) string {
	sign := 1
	if d < 0 {
		sign = -1
		d = -d
	}
	ns := int(d % 1e9)
	d /= 1e9
	sec := int(d % 60)
	d /= 60
	min := int(d % 60)
	hour := int(d/60) * sign
	if ns == 0 {
		return fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
	}
	return fmt.Sprintf("%d:%02d:%02d.%04d", hour, min, sec, ns/1e6)
}
