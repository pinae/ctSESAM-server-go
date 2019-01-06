/*
   Copyright (c) 2017-2018 Oliver Lau <ola@ct.de>, Heise Medien GmbH & Co. KG

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
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	realm           = "c't SESAM"
	version         = "0.1.2"
	port            = 8443
	indexReqURL     = "/index"
	listReqURL      = "/list"
	readReqURL      = "/read"
	writeReqURL     = "/write"
	credentialsFile = "./.htpasswd"
	deleteAfterDays = 90
	logFilename     = "SESAM.log"
	databaseFile    = "./ctsesam.sqlite.db"
	sqlCreateStmt   = "CREATE TABLE IF NOT EXISTS `domains` (" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"`userid` TEXT NOT NULL, " +
		"`created` INTEGER(4) DEFAULT (DATETIME('now', 'localtime')), " +
		"`data` BLOB " +
		");" +
		"CREATE INDEX IF NOT EXISTS `userid_idx` ON `domains` (`userid`);"
)

var (
	db *sql.DB
)

type _Page struct {
	Host string
	User string
}

func sendResponse(w http.ResponseWriter, result map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	if result["error"] != nil {
		result["status"] = "error"
	}
	var response []byte
	response, err := json.Marshal(result)
	if err != nil {
		log.Print(err)
	}
	w.Write(response)
}

func deleteOutdatedEntries(days int) (bool, string, int64) {
	sql := fmt.Sprintf(
		"DELETE FROM `domains` WHERE "+
			"`created` < DATETIME('now', 'localtime', '-%d days') AND "+
			"`created` != (SELECT MAX(`created`) FROM `domains`)",
		days)
	res, err := db.Exec(sql)
	if err != nil {
		return false, err.Error(), 0
	}
	rowsAffected, _ := res.RowsAffected()
	return true, "ok", rowsAffected
}

func cleanupJob(quitChannel chan bool) {
	for doQuit := false; !doQuit; {
		select {
		case doQuit = <-quitChannel:
		default:
		}
		ok, msg, rowsAffected := deleteOutdatedEntries(deleteAfterDays)
		if ok {
			log.Printf("Deleted %d outdated entries.", rowsAffected)
		} else {
			log.Printf("Deleting outdated entries failed: %s", msg)
		}
		time.Sleep(12 * time.Hour)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	log.Printf("%s request from %s (%s)", indexReqURL, user, r.Host)
	p := &_Page{r.Host, user}
	t, _ := template.ParseFiles("templates/default.tpl.html")
	t.Execute(w, p)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	log.Printf("%s request from %s (%s)", readReqURL, user, r.Host)
	result := make(map[string]interface{})
	stmt, err := db.Prepare(
		"SELECT `data` FROM `domains` " +
			"WHERE `userid` = ? " +
			"ORDER BY `created` DESC " +
			"LIMIT 1")
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
	log.Printf("%s request from %s (%s)", writeReqURL, user, r.Host)
	result := make(map[string]interface{})
	if r.Method == "POST" {
		data := strings.Replace(r.FormValue("data"), " ", "+", -1)
		stmt, err := db.Prepare(
			"INSERT INTO `domains` (userid, data) VALUES(?, ?)")
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
	log.Printf("%s request from %s (%s)", listReqURL, user, r.Host)
	result := make(map[string]interface{})
	stmt, err := db.Prepare(
		"SELECT `id`, `created`, LENGTH(`data`) " +
			"FROM `domains` " +
			"WHERE `userid` = ? " +
			"ORDER BY `created`")
	if err != nil {
		result["error"] = err.Error()
	} else {
		rows, err := stmt.Query(user)
		if err != nil {
			result["error"] = err.Error()
		} else {
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
	}
	if result["error"] != nil {
		result["status"] = "error"
	}
	sendResponse(w, result)
}

func main() {
	fmt.Printf("*** c't SESAM storage server %s (%s)\n", version, runtime.Version())
	fmt.Println("Copyright (c) 2017-2018 Oliver Lau <ola@ct.de>")
	fmt.Println("All rights reserved.")
	fmt.Println()

	fmt.Printf("Opening log file %s ...\n", logFilename)
	f, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	fmt.Printf("Parsing credentials in %s ...\n", credentialsFile)
	htpasswdFile, err := os.Open(credentialsFile)
	if err != nil {
		log.Fatal(err)
	}
	entries, err := newHTPasswd(htpasswdFile)
	htpasswdFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Opening database %s ...\n", databaseFile)
	db, err = sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(sqlCreateStmt)
	if err != nil {
		log.Fatalf("error creating database: %v", err)
	}
	defer db.Close()

	fmt.Println("Starting database cleanup job ...")
	quitChannel := make(chan bool)
	go cleanupJob(quitChannel)

	fmt.Printf("Starting secure web server on port %d ...\n", port)
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
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	mux.HandleFunc(readReqURL, auth(readHandler, entries, realm))
	mux.HandleFunc(listReqURL, auth(listHandler, entries, realm))
	mux.HandleFunc(writeReqURL, auth(writeHandler, entries, realm))
	mux.HandleFunc(indexReqURL, auth(indexHandler, entries, realm))

	intrChan := make(chan os.Signal, 1)
	signal.Notify(intrChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-intrChan
		log.Printf("Captured %v signal.", sig)
		srv.Shutdown(nil)
	}()
	log.Println("Starting.")
	err = srv.ListenAndServeTLS("cert/server.crt", "cert/server.key")
	if err != nil {
		fmt.Printf("ListenAndServeTLS() failed: %s\n", err.Error())
	}
	log.Println("Ending.")
}
