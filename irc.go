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
		joined := make(map[string]bool)
		for {
			msg := <-ch
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
