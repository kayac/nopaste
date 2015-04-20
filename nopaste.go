package nopaste

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/crowdmob/goamz/sns"
)

const Root = "/np"

var config *Config

type nopasteContent struct {
	Text    string
	Channel string
	Summary string
	Nick    string
	Notice  string
}

func Run(configFile string) error {
	var err error
	config, err = LoadConfig(configFile)
	if err != nil {
		return err
	}
	ch := make(chan IRCMessage, 64)
	go RunIRCAgent(config, ch)
	http.HandleFunc(Root, func(w http.ResponseWriter, req *http.Request) {
		rootHandler(w, req, ch)
	})
	http.HandleFunc(Root+"/", func(w http.ResponseWriter, req *http.Request) {
		serveHandler(w, req, ch)
	})
	http.HandleFunc(Root+"/amazon-sns/", func(w http.ResponseWriter, req *http.Request) {
		snsHandler(w, req, ch)
	})
	log.Fatal(http.ListenAndServe(config.Listen, nil))
	return nil
}

func rootHandler(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	if req.Method == "POST" {
		np := nopasteContent{
			Text:    req.FormValue("text"),
			Summary: req.FormValue("summary"),
			Notice:  req.FormValue("notice"),
			Channel: req.FormValue("channel"),
			Nick:    req.FormValue("nick"),
		}
		path, code := saveContent(np, ch)
		if code == http.StatusFound {
			http.Redirect(w, req, path, code)
		} else {
			serverError(w, code)
		}
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index", config.IRC); err != nil {
		log.Println(err)
		serverError(w, 500)
	}
}

func serveHandler(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	p := strings.Split(req.URL.Path, "/")
	if len(p) != 3 {
		http.NotFound(w, req)
		return
	}
	id := p[len(p)-1]
	if id == "" {
		rootHandler(w, req, ch)
		return
	}
	f, err := os.Open(config.DataFilePath(id))
	if err != nil {
		log.Println(err)
		http.NotFound(w, req)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	io.Copy(w, f)
}

func saveContent(np nopasteContent, ch chan IRCMessage) (string, int) {
	if np.Text == "" {
		return Root, http.StatusFound
	}
	data := []byte(np.Text)
	hex := fmt.Sprintf("%x", md5.Sum(data))
	id := hex[0:10]
	log.Println("save", id)
	err := ioutil.WriteFile(config.DataFilePath(id), data, 0644)
	if err != nil {
		log.Println(err)
		return Root, 500
	}
	if strings.Index(np.Channel, "#") == 0 {
		// post to irc
		summary := np.Summary
		nick := np.Nick
		url := config.BaseURL + Root + "/" + id
		msg := IRCMessage{
			Channel: np.Channel,
			Text:    fmt.Sprintf("%s %s %s", nick, summary, url),
			Notice:  false,
		}
		if np.Notice != "" {
			// true if 'notice' argument has any value (includes '0', 'false', 'null'...)
			msg.Notice = true
		}
		select {
		case ch <- msg:
		default:
			log.Println("Can't send msg to IRC")
		}
	}
	return Root + "/" + id, http.StatusFound
}

func serverError(w http.ResponseWriter, code int) {
	if code == 0 {
		code = http.StatusInternalServerError
	}
	http.Error(w, http.StatusText(code), code)
}

func snsHandler(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	if req.Method != "POST" {
		serverError(w, 400)
		return
	}
	var n *sns.HttpNotification
	dec := json.NewDecoder(req.Body)
	dec.Decode(&n)
	log.Println("sns", n.Type, n.TopicArn, n.Subject)
	switch n.Type {
	case "SubscriptionConfirmation", "Notification":
		if n.Type == "SubscriptionConfirmation" {
			s := NewSNS()
			_, err := s.ConfirmSubscriptionFromHttp(n, "no")
			if err != nil {
				log.Println(err)
				break
			}
		}
		p := strings.Split(req.URL.Path, "/")
		channel := p[len(p)-1]
		if channel == "" {
			break
		}
		var out bytes.Buffer
		json.Indent(&out, []byte(n.Message), "", "  ")
		np := nopasteContent{
			Text:    n.Message,
			Summary: out.String(),
			Notice:  "",
			Channel: "#" + channel,
		}
		saveContent(np, ch)
	}
	io.WriteString(w, "OK")
}
