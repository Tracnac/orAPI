package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"

	go_ora "github.com/sijms/go-ora/v2"
)

func process(w http.ResponseWriter, req *http.Request) {

	t, err := template.ParseFiles("templates" + req.URL.Path + ".sql")
	if err != nil {
		log.Fatal(err)
		return
	}

	u, err := url.Parse(req.RequestURI)
	if err != nil {
		log.Fatal(err)
		return
	}

	values := u.Query()

	var buf bytes.Buffer
	err = t.Execute(&buf, values)
	if err != nil {
		log.Fatal(err)
		return
	}

	js := oraConnect(buf.String())

	fmt.Fprintf(w, "%s\n", js)
}

func main() {
	http.HandleFunc("/", process)
	http.ListenAndServe(":8090", nil)
	os.Exit(0)
}

func oraConnect(query string) string {
	DBConStr := "oracle://tracnac:tracnac@192.168.122.90:1521/XEPDB1"
	DB, err := go_ora.NewConnection(DBConStr)
	checkErrExit("Driver error: ", err)
	err = DB.Open()
	checkErrExit("Open connection error: ", err)
	defer func() {
		err = DB.Close()
		checkErrExit("Connection close error: ", err)
	}()

	stmt := go_ora.NewStmt(query, DB)
	defer func() {
		err = stmt.Close()
		checkErrExit("Statement close error: ", err)
	}()

	rows, err := stmt.Query_(nil)
	checkErrExit("Query error: ", err)
	defer func() {
		err = rows.Close()
		checkErrExit("Cursor close error: ", err)
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
			str, err := json.Marshal(v)
			checkErrExit("(robot) Marshall Error", err)
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
	return tmp
}

func checkErrExit(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, msg, err)
		os.Exit(1)
	}
}
