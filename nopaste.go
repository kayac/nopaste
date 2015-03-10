package nopaste

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const Root = "/np"

var config *Config

func Run(configFile string) error {
	var err error
	config, err = LoadConfig(configFile)
	if err != nil {
		return err
	}
	ch := make(chan IRCMessage)
	go RunIRCAgent(config, ch)
	http.HandleFunc(Root, func(w http.ResponseWriter, req *http.Request) {
		rootHandler(w, req, ch)
	})
	http.HandleFunc(Root+"/", serveHandler)
	log.Fatal(http.ListenAndServe(config.Listen, nil))
	return nil
}

func rootHandler(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	if req.Method == "POST" {
		saveContent(w, req, ch)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index", config.IRC); err != nil {
		serverError(w, err)
	}
}

func serveHandler(w http.ResponseWriter, req *http.Request) {
	p := strings.Split(req.URL.Path, "/")
	if len(p) != 3 {
		http.NotFound(w, req)
		return
	}
	id := p[len(p)-1]
	f, err := os.Open(config.DataDir + "/" + id)
	if err != nil {
		log.Println(err)
		http.NotFound(w, req)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	io.Copy(w, f)
}

func saveContent(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	if req.FormValue("text") == "" {
		http.Redirect(w, req, Root, http.StatusFound)
		return
	}
	data := []byte(req.FormValue("text"))
	hex := fmt.Sprintf("%x", md5.Sum(data))
	id := hex[0:10]
	log.Println("save", id)
	err := ioutil.WriteFile(config.DataDir+"/"+id, data, 0644)
	if err != nil {
		log.Println(err)
		serverError(w, err)
		return
	}
	if channel := req.FormValue("channel"); channel != "" {
		// post to irc
		summary := req.FormValue("summary")
		nick := req.FormValue("nick")
		notice, _ := strconv.ParseBool(req.FormValue("notice"))
		url := config.BaseURL + Root + "/" + id
		msg := IRCMessage{
			Channel: channel,
			Text:    fmt.Sprintf("%s %s %s", nick, summary, url),
			Notice:  notice,
		}
		ch <- msg
	}
	http.Redirect(w, req, Root+"/"+id, http.StatusFound)
	return
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err)
	code := http.StatusInternalServerError
	http.Error(w, http.StatusText(code), code)
}
