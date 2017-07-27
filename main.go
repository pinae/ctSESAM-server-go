/*
   Copyright (c) 2017 Oliver Lau <ola@ct.de>, Heise Medien GmbH & Co. KG

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	_ "github.com/abbot/go-http-auth"
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
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{Host: r.Host}
	t, _ := template.ParseFiles("templates/default.tpl.html")
	t.Execute(w, p)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println(fmt.Sprintf("Starting web server on port %d ...", Port))
	http.HandleFunc("/", auth(indexHandler, entries))
	http.HandleFunc("/read", auth(readHandler, entries))
	http.HandleFunc("/write", auth(writeHandler, entries))
	http.HandleFunc("/delete", auth(deleteHandler, entries))
	http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
}
