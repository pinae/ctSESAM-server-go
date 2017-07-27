package main

import (
	"bufio"
	"errors"
	"fmt"
	_ "github.com/abbot/go-http-auth"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"strings"
)

type HTPasswd struct {
	entries map[string][]byte
}

func newHTPasswd(rd io.Reader) (*HTPasswd, error) {
	entries, err := parseHTPasswd(rd)
	if err != nil {
		return nil, err
	}
	return &HTPasswd{entries: entries}, nil
}

func (htpasswd *HTPasswd) authenticateUser(username string, password string) error {
	credentials, ok := htpasswd.entries[username]
	if !ok {
		// timing attack paranoia
		bcrypt.CompareHashAndPassword([]byte{}, []byte(password))
		return errors.New("authentication failure")
	}
	err := bcrypt.CompareHashAndPassword([]byte(credentials), []byte(password))
	if err != nil {
		return errors.New("authentication failure")
	}
	return nil
}

func parseHTPasswd(rd io.Reader) (map[string][]byte, error) {
	entries := map[string][]byte{}
	scanner := bufio.NewScanner(rd)
	line := 1
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if len(t) < 1 {
			continue
		}
		// lines that *begin* with a '#' are considered comments
		if t[0] == '#' {
			continue
		}
		i := strings.Index(t, ":")
		if i < 0 || i >= len(t) {
			return nil, fmt.Errorf("htpasswd: invalid entry at line %d: %q", line, scanner.Text())
		}
		user := t[:i]
		pass := t[i+1:]
		entries[user] = []byte(pass)
		line++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func auth(handler http.HandlerFunc, credentials *HTPasswd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		err := credentials.authenticateUser(user, pass)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+Realm+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n"))
			return
		}
		handler(w, r)
	}
}
