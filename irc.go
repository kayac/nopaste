package nopaste

import (
	"fmt"
	"log"
	"time"

	irc "github.com/thoj/go-ircevent"
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
	joined := make(map[string]bool)
	for {
		select {
		case <-done:
			return
		case msg := <-ch:
			if !joined[msg.Channel] {
				log.Println("join", msg.Channel)
				agent.Join(msg.Channel)
				joined[msg.Channel] = true
				c.AddChannel(msg.Channel)
			}
			if msg.Notice {
				agent.Notice(msg.Channel, msg.Text)
			} else {
				agent.Privmsg(msg.Channel, msg.Text)
			}
		}
	}
}
