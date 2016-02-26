// Command
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GeertJohan/go.rice"
	"github.com/tmc/adt"
)

type Server struct {
	cfg     Config
	path    string
	verbose bool
}

func NewADTHTTPServer(cfg Config, path string, verbose bool) *Server {
	return &Server{cfg: cfg, path: path, verbose: verbose}
}

func (s *Server) Serve(addr string, publicKeyPath string, privateKeyPath string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.srvIndex)
	mux.HandleFunc("/dbs/", s.srvDBs)
	if publicKeyPath != "" && privateKeyPath != "" {
		return http.ListenAndServeTLS(*flagAddr, publicKeyPath, privateKeyPath, mux)
	}
	return http.ListenAndServe(*flagAddr, mux)
}

func (s *Server) srvIndex(rw http.ResponseWriter, r *http.Request) {
	render(rw, "index.tmpl", listdbs())
}

func (s *Server) srvDBs(rw http.ResponseWriter, r *http.Request) {
	cors(rw, r)
	name := r.URL.Path[len("/dbs/"):]
	parts := strings.Split(name, "/")
	switch len(parts) {
	case 1:
		s.srvDBIndex(rw, r)
	case 2:
		s.srvDBRecord(rw, r)
	}
}

func (s *Server) srvDBIndex(rw http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/dbs/"):]
	parts := strings.Split(name, "/")
	table, err := adt.TableFromPath(filepath.Join(s.path, parts[0]))
	if renderErr(rw, err) {
		return
	}
	if query := r.URL.Query().Get("q"); query != "" {
		rw.Header().Add("Content-Type", "application/json")
		field := r.URL.Query().Get("field")
		for i := int(table.RecordCount) - 1; i >= 0; i-- {
			data, err := table.Get(i)
			if renderErr(rw, err) {
				return
			}
			if fmt.Sprint(data[field]) == query {
				data, err = s.decorateRecord(parts[0], data)
				if renderErr(rw, err) {
					return
				}
				json.NewEncoder(rw).Encode(data)
				return
			}
		}
		rw.WriteHeader(http.StatusNotFound)
	}
	render(rw, "db.tmpl", table)
}

func (s *Server) srvDBRecord(rw http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/dbs/"):]
	parts := strings.Split(name, "/")
	table, err := adt.TableFromPath(filepath.Join(s.path, parts[0]))
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
	data, err = s.decorateRecord(parts[0], data)
	if renderErr(rw, err) {
		return
	}
	rw.Header().Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(data)
}

func (s *Server) decorateRecord(table string, record map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	tblConf, hasConf := s.cfg[Table(table)]

	for key, value := range record {
		if hasConf && tblConf[Column(key)] != "" {
			related, err := s.lookupRecord(tblConf[Column(key)], value)
			if err != nil {
				log.Println("issue looking up related record", table, key, value)
			} else {
				value = related
			}
		}
		result[key] = value
	}
	return result, nil
}

func (s *Server) lookupRecord(tableName Table, pk interface{}) (map[string]interface{}, error) {
	table, err := adt.TableFromPath(filepath.Join(s.path, string(tableName)))
	if err != nil {
		return nil, err
	}
	pkCol, err := table.GetPK()
	if err != nil {
		return nil, err
	}
	for i := int(table.RecordCount) - 1; i >= 0; i-- {
		record, err := table.Get(i)
		if err != nil {
			return nil, err
		}
		if fmt.Sprint(record[pkCol.Name]) == fmt.Sprint(pk) {
			return record, nil
		}
	}
	return nil, nil
}

func cors(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
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
