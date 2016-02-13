// Command
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/tmc/adt"
)

var (
	flagFile  = flag.String("f", "", "path to ADT file")
	flagIndex = flag.Int("i", 0, "starting index")
	flagNum   = flag.Int("n", -1, "number of records")
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

	w := csv.NewWriter(os.Stdout)
	// header
	headers := []string{}
	for _, c := range table.Columns {
		headers = append(headers, c.Name)
	}
	if err := w.Write(headers); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for i := *flagIndex; i < until; i++ {
		fields := make([]string, 0, len(table.Columns))
		r, err := table.Get(i)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, c := range table.Columns {
			fields = append(fields, fmt.Sprint(r[c.Name]))
		}
		if err := w.Write(fields); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	w.Flush()
}
