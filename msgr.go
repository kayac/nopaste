package nopaste

import (
	"context"
	"log"
	"net/http"
)

const MsgrRoot = "/irc-msgr"

func RunMsgr(ctx context.Context, configFile string) error {
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
	http.HandleFunc(MsgrRoot+"/post", func(w http.ResponseWriter, req *http.Request) {
		msgrPostHandler(w, req, chs)
	})
	log.Fatal(http.ListenAndServe(config.Listen, nil))
	return nil
}

func msgrPostHandler(w http.ResponseWriter, req *http.Request, chs []MessageChan) {
	channel := req.FormValue("channel")
	msg := req.FormValue("msg")
	if channel == "" || msg == "" || req.Method != "POST" {
		log.Printf("[warn] invalid request: method=%s channel=%s, msg=%s", req.Method, channel, msg)
		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
		return
	}
	log.Printf("[debug] post to channel=%s, msg=%s", channel, msg)
	for _, ch := range chs {
		ch.PostMsgr(req)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte{})
}
