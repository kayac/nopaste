package nopaste

import (
	"fmt"
	"log"
	"time"

	irc "github.com/thoj/go-ircevent"
)

const (
	MsgBufferLen = 100
)

var (
	IRCThrottleWindow = 1 * time.Second
)

type IRCMessage struct {
	Channel string
	Notice  bool
	Text    string
}

func RunIRCAgent(c *Config, ch chan IRCMessage) {
	for {
		agent := irc.IRC(c.IRC.Nick, c.IRC.Nick)
		agent.UseTLS = c.IRC.Secure
		agent.Password = c.IRC.Password
		addr := fmt.Sprintf("%s:%d", c.IRC.Host, c.IRC.Port)
		err := agent.Connect(addr)
		if err != nil {
			log.Println(err)
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
				log.Println("join", msg.Channel)
				agent.Join(msg.Channel)
				joined[msg.Channel] = make(chan IRCMessage, MsgBufferLen)
				go sendMsgToIRCChannel(agent, joined[msg.Channel], done)
				c.AddChannel(msg.Channel)
			}
			select {
			case joined[msg.Channel] <- msg:
			default:
				log.Println("Can't send msg to IRC. Channel buffer flooding.")
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
		log.Println("throttled. sleeping", window-diff)
		time.Sleep(window - diff)
	}
}
