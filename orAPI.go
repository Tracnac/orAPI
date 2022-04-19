package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"text/template"

	go_ora "github.com/sijms/go-ora/v2"
)

func process(w http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" || req.URL.Path == "/favicon.ico" {
		return
	}

	t, err := template.ParseFiles("templates" + req.URL.Path + ".sql")
	if err != nil {
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	u, err := url.Parse(req.RequestURI)
	if err != nil {
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	values := u.Query()

	var buf bytes.Buffer
	err = t.Execute(&buf, values)
	if err != nil {
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	js := oraConnect(buf.String())

	fmt.Fprintf(w, "%s\n", js)
}

func main() {
	http.HandleFunc("/", process)
	err := http.ListenAndServe("0.0.0.0:8090", nil)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func oraConnect(query string) string {
	DBConStr := "oracle://tracnac:tracnac@192.168.122.90:1521/XEPDB1"
	DB, err := go_ora.NewConnection(DBConStr)
	if err != nil {
		return err.Error()
	}
	err = DB.Open()
	if err != nil {
		return err.Error()
	}
	defer func() {
		_ = DB.Close()
	}()

	stmt := go_ora.NewStmt(query, DB)
	defer func() {
		_ = stmt.Close()
	}()

	rows, err := stmt.Query_(nil)
	if err != nil {
		return err.Error()
	}
	defer func() {
		_ = rows.Close()
	}()

	var tmp string
	_len := len(rows.Columns()) - 1

	tmp = "[\n"

	first := true
	for rows.Next_() {
		if !first {
			tmp += "},\n  {"
		} else {
			first = false
			tmp += "  {"
		}
		for k, v := range rows.CurrentRow {
			str, _ := json.Marshal(v)
			if k < _len {
				tmp += fmt.Sprintf("\"%s\": %v, ", rows.Columns()[k], string(str))
			} else {
				tmp += fmt.Sprintf("\"%s\": %v", rows.Columns()[k], string(str))
			}
		}
	}
	if first {
		tmp += "]\n"
	} else {
		tmp += "}\n]\n"
	}
	return jsonPrettyPrint(tmp)
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "  ")
	if err != nil {
		return in
	}
	return out.String()
}
