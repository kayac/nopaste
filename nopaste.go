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
	Text      string
	Channel   string
	Summary   string
	Nick      string
	Notice    string
	IconEmoji string
	IconURL   string
	LinkNames int
}

func Run(configFile string) error {
	var err error
	config, err = LoadConfig(configFile)
	if err != nil {
		return err
	}
	var chs []MessageChan
	if config.IRC != nil {
		ircCh := make(IRCMessageChan, MsgBufferLen)
		chs = append(chs, ircCh)
		go RunIRCAgent(config, ircCh)
	}
	if config.Slack != nil {
		slackCh := make(SlackMessageChan, MsgBufferLen)
		chs = append(chs, slackCh)
		go RunSlackAgent(config, slackCh)
	}

	http.HandleFunc(Root, func(w http.ResponseWriter, req *http.Request) {
		rootHandler(w, req, chs)
	})
	http.HandleFunc(Root+"/", func(w http.ResponseWriter, req *http.Request) {
		serveHandler(w, req, chs)
	})
	http.HandleFunc(Root+"/amazon-sns/", func(w http.ResponseWriter, req *http.Request) {
		snsHandler(w, req, chs)
	})
	log.Fatal(http.ListenAndServe(config.Listen, nil))
	return nil
}

func rootHandler(w http.ResponseWriter, req *http.Request, chs []MessageChan) {
	if req.Method == "POST" {
		np := nopasteContent{
			Text:      req.FormValue("text"),
			Summary:   req.FormValue("summary"),
			Notice:    req.FormValue("notice"),
			Channel:   req.FormValue("channel"),
			Nick:      req.FormValue("nick"),
			IconEmoji: req.FormValue("icon_emoji"),
			IconURL:   req.FormValue("icon_url"),
			LinkNames: 0,
		}
		if _notice := req.FormValue("notice"); _notice == "0" {
			np.LinkNames = 1
		}
		path, code := saveContent(np, chs)
		if code == http.StatusFound {
			http.Redirect(w, req, path, code)
		} else {
			serverError(w, code)
		}
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index", config); err != nil {
		log.Println(err)
		serverError(w, 500)
	}
}

func serveHandler(w http.ResponseWriter, req *http.Request, chs []MessageChan) {
	p := strings.Split(req.URL.Path, "/")
	if len(p) != 3 {
		http.NotFound(w, req)
		return
	}
	id := p[len(p)-1]
	if id == "" {
		rootHandler(w, req, chs)
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

func saveContent(np nopasteContent, chs []MessageChan) (string, int) {
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
		url := config.BaseURL + Root + "/" + id
		for _, ch := range chs {
			ch.PostNopaste(np, url)
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

func snsHandler(w http.ResponseWriter, req *http.Request, chs []MessageChan) {
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
			region, _ := getRegionFromARN(n.TopicArn)
			s := NewSNS(region)
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
		out.WriteString(n.Type)
		out.WriteString(n.TopicArn)
		out.WriteString("\n")
		json.Indent(&out, []byte(n.Message), "", "  ")

		subject := n.Subject
		if key := req.FormValue("key"); key != "" {
			var m map[string]interface{}
			json.Unmarshal([]byte(n.Message), &m)
			if v, found := m[key]; found {
				subject = fmt.Sprintf("%s %s", subject, v)
			}
		}
		np := nopasteContent{
			Text:      out.String(),
			Summary:   subject,
			Notice:    "",
			Channel:   "#" + channel,
			IconEmoji: ":amazonsns:",
			Nick:      "AmazonSNS",
		}
		saveContent(np, chs)
	}
	io.WriteString(w, "OK")
}
