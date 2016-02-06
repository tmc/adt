package adt_test

import (
	"os"
	"testing"

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
	for i := 0; i < int(table.RecordCount); i++ {
		_, err := table.Get(i)
		//fmt.Printf("%+v\n", r)
		if err != nil {
			t.Error("row", i, err)
		}
	}
}
