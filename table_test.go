package adt_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/tmc/adt"
)

func TestTableRead(t *testing.T) {
	db := os.Getenv("ADT_TEST_FILE")
	if db == "" {
		t.Skip("ADT_TEST_FILE not set")
	}

	table, err := adt.TableFromPath(db)
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(table)
	r, err := table.Get(int(table.RecordCount - 2))
	fmt.Printf("%+v\n", r)
}
