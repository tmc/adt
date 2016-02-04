// Command adtdump dumps information about an ADT database
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/tmc/adt"
)

var flagFile = flag.String("f", "", "path to ADT file")

func main() {
	flag.Parse()
	table, err := adt.TableFromPath(*flagFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	spew.Dump(table)
	r, err := table.Get(int(table.RecordCount - 2))
	fmt.Printf("%+v\n", r)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
