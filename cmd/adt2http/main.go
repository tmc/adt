// Command
package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/tmc/adt"
)

var (
	flagPath    = flag.String("path", ".", "path to ADT files")
	flagVerbose = flag.Bool("v", false, "verbose")
	flagAddr    = flag.String("http", ":7000", "listen address")
)

func main() {
	flag.Parse()
	if err := serve(); err != nil {
		log.Fatalln(err)
	}
}

func serve() error {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(rice.MustFindBox("static").HTTPBox())))
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		render(rw, "index.tmpl", listdbs())
	})
	mux.HandleFunc("/dbs/", func(rw http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/dbs/"):]
		parts := strings.Split(name, "/")
		switch len(parts) {
		case 1:
			table, err := adt.TableFromPath(filepath.Join(*flagPath, parts[0]))
			if renderErr(rw, err) {
				return
			}
			render(rw, "db.tmpl", table)
		case 2:
			table, err := adt.TableFromPath(filepath.Join(*flagPath, parts[0]))
			if renderErr(rw, err) {
				return
			}
			index, err := strconv.Atoi(parts[1])
			if renderErr(rw, err) {
				return
			}
			data, err := table.Get(index - 1)
			if renderErr(rw, err) {
				return
			}
			rw.Header().Add("Content-Type", "application/json")
			json.NewEncoder(rw).Encode(data)
		}
	})

	return http.ListenAndServe(*flagAddr, mux)
}

func render(rw http.ResponseWriter, tmpl string, data interface{}) {
	t, err := getTmpl(tmpl)
	if renderErr(rw, err) {
		return
	}
	t.Execute(rw, data)
}

func renderErr(rw http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	rw.WriteHeader(500)
	rw.Write([]byte(err.Error()))
	return true
}

func getTmpl(name string) (*template.Template, error) {
	templates := rice.MustFindBox("templates")
	t, err := template.New("").Parse(templates.MustString("base.tmpl"))
	if err != nil {
		return nil, err
	}
	tmpl, err := templates.String(name)
	if err != nil {
		return nil, err
	}
	return t.Parse(tmpl)
}

func listdbs() []string {
	result, err := filepath.Glob(filepath.Join(*flagPath, "*.ADT"))
	if err != nil {
		panic(err)
	}
	for i, s := range result {
		result[i] = filepath.Base(s)
	}
	return result
}
