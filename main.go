/*
   Copyright (c) Oliver Lau <ola@ct.de>, Heise Medien GmbH & Co. KG

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	Realm           = "c't SESAM"
	Version         = "0.0.1-DEV"
	Port            = 8088
	CredentialsFile = "./.htpasswd"
	DatabaseFile    = "./ctsesam.sqlite.db"
)

var (
	db *sql.DB = nil
)

type Page struct {
	Host string
	User string
}

func sendResponse(w http.ResponseWriter, result map[string]interface{}) {
	if result["error"] != nil {
		result["status"] = "error"
	}
	if len(result) > 0 {
		w.Header().Add("Content-Type", "application/json")
		var response []byte
		response, err := json.Marshal(result)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(response)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	p := &Page{r.Host, user}
	t, _ := template.ParseFiles("templates/default.tpl.html")
	t.Execute(w, p)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	result := make(map[string]interface{})
	stmt, err := db.Prepare("SELECT `data` FROM `domains` WHERE `userid` = ? ORDER BY `created` DESC LIMIT 1")
	if err != nil {
		result["error"] = err.Error()
	}
	var data []byte
	err = stmt.QueryRow(user).Scan(&data)
	switch {
	case err == sql.ErrNoRows:
		result["status"] = "ok"
		result["info"] = "No user with that ID."
	case err != nil:
		result["error"] = err.Error()
	default:
		result["result"] = string(data)
		result["status"] = "ok"
	}
	if result["error"] != nil {
		result["status"] = "error"
	}
	sendResponse(w, result)
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	result := make(map[string]interface{})
	if r.Method == "POST" {
		data := strings.Replace(r.FormValue("data"), " ", "+", -1)
		stmt, err := db.Prepare("INSERT INTO `domains` (userid, data) VALUES(?, ?)")
		_, err = stmt.Exec(user, data)
		if err != nil {
			result["error"] = err.Error()
		}
	} else {
		http.Error(w, "Invalid request method.", 405)
	}
	sendResponse(w, result)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	result := make(map[string]interface{})
	stmt, err := db.Prepare("SELECT `id`, `created`, LENGTH(`data`) FROM `domains` WHERE `userid` = ? ORDER BY `created`")
	if err != nil {
		result["error"] = err.Error()
	}
	rows, err := stmt.Query(user)
	switch {
	case err != nil:
		result["error"] = err.Error()
	default:
		res := make([]map[string]interface{}, 0)
		for rows.Next() {
			item := make(map[string]interface{})
			var id int
			var created string
			var sz int
			err := rows.Scan(&id, &created, &sz)
			if err != nil {
				log.Fatal(err)
			} else {
				item["id"] = id
				item["created"] = created
				item["size"] = sz
				res = append(res, item)
			}
		}
		result["result"] = res
		result["status"] = "ok"
	}
	if result["error"] != nil {
		result["status"] = "error"
	}
	sendResponse(w, result)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	result := make(map[string]interface{})
	result["action"] = "not implemented yet"
	sendResponse(w, result)
}

func main() {
	fmt.Println(fmt.Sprintf("*** c't SESAM storage server %s ***", Version))
	fmt.Println(fmt.Sprintf("Parsing credentials in %s ...", CredentialsFile))
	htpasswd_file, err := os.Open(CredentialsFile)
	if err != nil {
		log.Fatal(err)
	}
	entries, err := newHTPasswd(htpasswd_file)
	htpasswd_file.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Opening database %s ...", DatabaseFile))
	db, err = sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Starting secure web server on port %d ...", Port))
	mux := http.NewServeMux()
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS10,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", Port),
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	mux.HandleFunc("/", auth(indexHandler, entries, Realm))
	mux.HandleFunc("/read", auth(readHandler, entries, Realm))
	mux.HandleFunc("/list", auth(listHandler, entries, Realm))
	mux.HandleFunc("/write", auth(writeHandler, entries, Realm))
	mux.HandleFunc("/delete", auth(deleteHandler, entries, Realm))
	srv.ListenAndServeTLS("cert/server.crt", "cert/private/server.key")
	db.Close()
}
