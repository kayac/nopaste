package nopaste

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

const Root = "/np"

var config *Config

var Debug = false

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

func Run(ctx context.Context, configFile string) error {
	var err error
	config, err = LoadConfig(ctx, configFile)
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
		path, code := saveContent(req.Context(), np, chs)
		if code == http.StatusFound {
			http.Redirect(w, req, path, code)
		} else {
			serverError(w, code)
		}
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index", config); err != nil {
		log.Println("[warn]", err)
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

	var f io.ReadCloser
	var err error
	for _, s := range config.Storages() {
		if _f, _err := s.Load(req.Context(), id); _err == nil {
			log.Println("[debug] loaded from", s)
			f, err = _f, _err
			break
		} else {
			err = _err
		}
	}
	if err != nil || f == nil {
		log.Println("[warn]", err)
		http.NotFound(w, req)
		return
	}
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	io.Copy(w, f)
}

func saveContent(ctx context.Context, np nopasteContent, chs []MessageChan) (string, int) {
	if np.Text == "" {
		log.Println("[warn] empty text")
		return Root, http.StatusFound
	}
	data := []byte(np.Text)
	hex := fmt.Sprintf("%x", md5.Sum(data))
	id := hex[0:10]
	log.Println("[info] save", id)

	err := config.Storages()[0].Save(ctx, id, data)
	if err != nil {
		log.Println("[warn]", err)
		return Root, 500
	}
	if strings.Index(np.Channel, "#") == 0 {
		url := config.BaseURL + Root + "/" + id
		for _, ch := range chs {
			ch.PostNopaste(np, url)
		}
	}
	return path.Join(Root, id), http.StatusFound
}

func serverError(w http.ResponseWriter, code int) {
	if code == 0 {
		code = http.StatusInternalServerError
	}
	http.Error(w, http.StatusText(code), code)
}

// https://docs.aws.amazon.com/sns/latest/dg/json-formats.html
type HttpNotification struct {
	Type             string    `json:"Type"`
	MessageId        string    `json:"MessageId"`
	Token            string    `json:"Token"` // Only for subscribe and unsubscribe
	TopicArn         string    `json:"TopicArn"`
	Subject          string    `json:"Subject"` // Only for Notification
	Message          string    `json:"Message"`
	SubscribeURL     string    `json:"SubscribeURL"` // Only for subscribe and unsubscribe
	Timestamp        time.Time `json:"Timestamp"`
	SignatureVersion string    `json:"SignatureVersion"`
	Signature        string    `json:"Signature"`
	SigningCertURL   string    `json:"SigningCertURL"`
	UnsubscribeURL   string    `json:"UnsubscribeURL"` // Only for notifications
}

func snsHandler(w http.ResponseWriter, req *http.Request, chs []MessageChan) {
	ctx := req.Context()
	if req.Method != "POST" {
		serverError(w, 400)
		return
	}
	p := strings.Split(req.URL.Path, "/")
	channel := p[len(p)-1]
	if channel == "" {
		serverError(w, 404)
		return
	}
	var n HttpNotification
	b, _ := io.ReadAll(req.Body)
	if err := json.Unmarshal(b, &n); err != nil {
		log.Println("[warn]", err)
	}
	log.Println("[info] sns", n.Type, n.TopicArn, n.Subject)
	switch n.Type {
	case "":
		// raw message
		np := nopasteContent{
			Text:      string(b),
			Summary:   "Received raw message",
			Notice:    "",
			Channel:   "#" + channel,
			IconEmoji: ":amazonsns:",
			Nick:      "AmazonSNS",
		}
		saveContent(ctx, np, chs)
	case "SubscriptionConfirmation", "Notification":
		if n.Type == "SubscriptionConfirmation" {
			region, _ := getRegionFromARN(n.TopicArn)
			snsSvc, err := NewSNS(ctx, region)
			if err != nil {
				log.Println("[warn]", err)
				break
			}
			if _, err := snsSvc.ConfirmSubscription(ctx, &sns.ConfirmSubscriptionInput{
				Token:                     aws.String(n.Token),
				TopicArn:                  aws.String(n.TopicArn),
				AuthenticateOnUnsubscribe: aws.String("no"),
			}); err != nil {
				log.Println("[warn]", err)
				break
			}
		}
		var out bytes.Buffer
		fmt.Fprintf(&out, "%s from %s\n", n.Type, n.TopicArn)
		if subscriptionArn := req.Header.Get("x-amz-sns-subscription-arn"); subscriptionArn != "" {
			fmt.Fprintf(&out, "Subscribe by %s\n", subscriptionArn)
		}
		if err := json.Indent(&out, []byte(n.Message), "", "  "); err != nil {
			out.WriteString(n.Message) // invalid JSON
		}

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
		saveContent(ctx, np, chs)
	}
	io.WriteString(w, "OK")
}
