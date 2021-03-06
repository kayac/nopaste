package nopaste

import (
	"fmt"
	"log"
	"net/http"
	"time"

	irc "github.com/thoj/go-ircevent"
)

const (
	MsgBufferLen = 100
)

var (
	IRCThrottleWindow = 1 * time.Second
)

type MessageChan interface {
	PostNopaste(np nopasteContent, url string)
	PostMsgr(np *http.Request)
}

type IRCMessage struct {
	Channel string
	Notice  bool
	Text    string
}

type IRCMessageChan chan IRCMessage

func (ch IRCMessageChan) PostNopaste(np nopasteContent, url string) {
	summary := np.Summary
	nick := np.Nick
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
		log.Println("[warn] Can't send msg to IRC")
	}
}

func (ch IRCMessageChan) PostMsgr(req *http.Request) {
	msg := IRCMessage{
		Channel: req.FormValue("channel"),
		Text:    req.FormValue("msg"),
		Notice:  true,
	}
	if _notice := req.FormValue("notice"); _notice == "" || _notice == "0" {
		msg.Notice = false
	}
	select {
	case ch <- msg:
	default:
		log.Println("[warn] Can't send msg to IRC")
	}
}

func RunIRCAgent(c *Config, ch chan IRCMessage) {
	log.Println("[info] running irc agent")
	for {
		agent := irc.IRC(c.IRC.Nick, c.IRC.Nick)
		agent.UseTLS = c.IRC.Secure
		agent.Password = c.IRC.Password
		addr := fmt.Sprintf("%s:%d", c.IRC.Host, c.IRC.Port)
		err := agent.Connect(addr)
		if err != nil {
			log.Println("[warn]", err)
			time.Sleep(10 * time.Second)
			continue
		}
		done := make(chan interface{})
		go sendMsgToIRC(c, agent, ch, done)
		agent.Loop()
		close(done)
		time.Sleep(10 * time.Second)
	}
}

func sendMsgToIRC(c *Config, agent *irc.Connection, ch chan IRCMessage, done chan interface{}) {
	joined := make(map[string]chan IRCMessage)
	for {
		select {
		case <-done:
			return
		case msg := <-ch:
			if _, ok := joined[msg.Channel]; !ok {
				log.Println("[info] join", msg.Channel)
				agent.Join(msg.Channel)
				joined[msg.Channel] = make(chan IRCMessage, MsgBufferLen)
				go sendMsgToIRCChannel(agent, joined[msg.Channel], done)
				c.AddChannel(msg.Channel)
			}
			select {
			case joined[msg.Channel] <- msg:
			default:
				log.Println("[warn] Can't send msg to IRC. Channel buffer flooding.")
			}
		}
	}
}

func sendMsgToIRCChannel(agent *irc.Connection, ch chan IRCMessage, done chan interface{}) {
	lastPostedAt := time.Now()
	for {
		select {
		case <-done:
			return
		case msg := <-ch:
			throttle(lastPostedAt, IRCThrottleWindow)
			if msg.Notice {
				agent.Notice(msg.Channel, msg.Text)
			} else {
				agent.Privmsg(msg.Channel, msg.Text)
			}
			lastPostedAt = time.Now()
		}
	}
}

func throttle(last time.Time, window time.Duration) {
	now := time.Now()
	diff := now.Sub(last)
	if diff < window {
		// throttle
		log.Println("[info] throttled. sleeping", window-diff)
		time.Sleep(window - diff)
	}
}
