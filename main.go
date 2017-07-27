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
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	Realm           = "c't SESAM"
	Port            = 8088
	CredentialsFile = ".htpasswd"
)

type Page struct {
	Host string
	User string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	user, _, _ := r.BasicAuth()
	p := &Page{Host: r.Host, User: user}
	t, _ := template.ParseFiles("templates/default.tpl.html")
	t.Execute(w, p)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	stmt, err := sql.Prepare("SELECT * FROM `domains` WHERE `userid` = ?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err = stmt.Query(user)
	if err != nil {
		log.Fatal(err)
	}
	var user string
	var blob []byte
	for rows.Next() {
		err = rows.Scan(&uid, &blob)
		fmt.Println(user, blob)
	}
	rows.Close()
	db.Close()
}

func writeHandler(w http.ResponseWriter, r *http.Request) {
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
}

func main() {
	fmt.Println("*** c't SESAM storage server ***")
	fmt.Println(fmt.Sprintf("Parsing %s ...", CredentialsFile))
	htpasswd_file, err := os.Open("./" + CredentialsFile)
	if err != nil {
		log.Fatal(err)
	}
	entries, err := newHTPasswd(htpasswd_file)
	htpasswd_file.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Starting secure web server on port %d ...", Port))
	mux := http.NewServeMux()
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
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
	mux.HandleFunc("/write", auth(writeHandler, entries, Realm))
	mux.HandleFunc("/delete", auth(deleteHandler, entries, Realm))
	srv.ListenAndServeTLS("cert/server.rsa.crt", "cert/server.rsa.key")
}
