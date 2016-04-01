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
	flagFile   = flag.String("f", "", "path to ADT file")
	flagIndex  = flag.Int("i", 0, "starting index")
	flagNum    = flag.Int("n", -1, "number of records")
	flagIndent = flag.Bool("indent", false, "ident")
)

func main() {
	flag.Parse()
	table, err := adt.TableFromPath(*flagFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	until := int(table.RecordCount)
	if *flagNum != -1 {
		until = *flagIndex + *flagNum
	}
	for i := *flagIndex; i < until; i++ {
		r, err := table.Get(i)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var buf []byte
		if *flagIndent {
			buf, _ = json.MarshalIndent(r, "", "  ")
		} else {
			buf, _ = json.Marshal(r)
		}
		os.Stdout.Write(buf)
		os.Stdout.WriteString("\n")
	}

}
