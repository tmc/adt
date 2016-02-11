package adt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var ddlTmpl = `CREATE TABLE IF NOT EXISTS "{{.TableName}}" (
{{range $index, $column := .Columns }}{{if $index}},{{end}}
	"{{.Name}}" {{.Type.SQLType}}{{end}}
)`

func (t *Table) SQLDDL(tableName string) (string, error) {
	tmpl, err := template.New("").Parse(ddlTmpl)
	if err != nil {
		return "", err
	}
	output := new(bytes.Buffer)
	err = tmpl.Execute(output, map[string]interface{}{
		"TableName": tableName,
		"Columns":   t.Columns,
	})
	return string(output.Bytes()), err
}

func (t *Table) InsertSQL(tableName string) string {
	placeHolders := make([]string, 0, len(t.Columns))
	for i := 0; i < len(t.Columns); i++ {
		placeHolders = append(placeHolders, "?") // todo(mysqlism)
	}
	result := fmt.Sprintf(`INSERT INTO "%s" VALUES(%s)`, tableName, strings.Join(placeHolders, ","))
	return result
}
