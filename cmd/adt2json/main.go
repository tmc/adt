// Command
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tmc/adt"
)

var (
	flagFile  = flag.String("f", "", "path to ADT file")
	flagIndex = flag.Int("i", 0, "starting index")
)

func main() {
	flag.Parse()
	table, err := adt.TableFromPath(*flagFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	for i := *flagIndex; i < int(table.RecordCount); i++ {
		r, err := table.Get(i)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		enc.Encode(r)
	}

}
