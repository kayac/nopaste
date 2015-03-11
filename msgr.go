package nopaste

import (
	"log"
	"net/http"
)

const MsgrRoot = "/irc-msgr"

func RunMsgr(configFile string) error {
	var err error
	config, err = LoadConfig(configFile)
	if err != nil {
		return err
	}
	ch := make(chan IRCMessage)
	go RunIRCAgent(config, ch)
	http.HandleFunc(MsgrRoot+"/post", func(w http.ResponseWriter, req *http.Request) {
		msgrPostHandler(w, req, ch)
	})
	log.Fatal(http.ListenAndServe(config.Listen, nil))
	return nil
}

func msgrPostHandler(w http.ResponseWriter, req *http.Request, ch chan IRCMessage) {
	channel := req.FormValue("channel")
	msg := req.FormValue("msg")
	if channel == "" || msg == "" || req.Method != "POST" {
		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
		return
	}
	ircMsg := IRCMessage{
		Channel: channel,
		Text:    msg,
		Notice:  true,
	}
	if _notice := req.FormValue("notice"); _notice == "" || _notice == "0" {
		ircMsg.Notice = false
	}
	log.Printf("%#v", ircMsg)
	ch <- ircMsg

	code := http.StatusCreated
	http.Error(w, http.StatusText(code), code)
}
