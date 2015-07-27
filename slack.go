package nopaste

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	SlackThrottleWindow = 1 * time.Second
)

type SlackMessage struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji"`
	Username  string `json:"username"`
}

type SlackMessageChan chan SlackMessage

func (ch SlackMessageChan) Post(np nopasteContent, url string) {
	summary := np.Summary
	nick := np.Nick
	var text string
	if summary == "" {
		text = fmt.Sprintf("<%s|%s>", url, url)
	} else {
		text = fmt.Sprintf("%s <%s|show details>", summary, url)
	}
	msg := SlackMessage{
		Channel:   np.Channel,
		Username:  nick,
		Text:      text,
		IconEmoji: np.IconEmoji,
	}
	select {
	case ch <- msg:
	default:
		log.Println("Can't send msg to Slack")
	}
}

type SlackAgent struct {
	WebhookURL string
	client     *http.Client
}

func (a *SlackAgent) Post(m SlackMessage) {
	payload, _ := json.Marshal(&m)
	v := url.Values{}
	v.Set("payload", string(payload))
	log.Println("post to slack", a, string(payload))
	resp, err := a.client.PostForm(a.WebhookURL, v)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Println("failed post to slack:", body)
	} else {
		log.Println(err)
	}
}

func RunSlackAgent(c *Config, ch chan SlackMessage) {
	log.Println("runing slack agent")
	joined := make(map[string]chan SlackMessage)
	agent := &SlackAgent{
		WebhookURL: c.Slack.WebhookURL,
		client:     &http.Client{},
	}
	for {
		select {
		case msg := <-ch:
			if _, ok := joined[msg.Channel]; !ok {
				joined[msg.Channel] = make(chan SlackMessage, MsgBufferLen)
				go sendMsgToSlackChannel(agent, joined[msg.Channel])
			}
			select {
			case joined[msg.Channel] <- msg:
			default:
				log.Println("Can't send msg to Slack. Channel buffer flooding.")
			}
		}
	}
}

func sendMsgToSlackChannel(agent *SlackAgent, ch chan SlackMessage) {
	lastPostedAt := time.Now()
	for {
		select {
		case msg := <-ch:
			throttle(lastPostedAt, SlackThrottleWindow)
			agent.Post(msg)
			lastPostedAt = time.Now()
		}
	}
}
