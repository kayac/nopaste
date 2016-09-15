package nopaste

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

const SlackMaxBackOff = 3600

var (
	SlackThrottleWindow = 1 * time.Second
	SlackInitialBackOff = 30
	Epoch               = time.Unix(0, 0)
)

type SlackMessage struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Username  string `json:"username"`
	LinkNames int    `json:"link_names,omitempty"`
}

type SlackMessageChan chan SlackMessage

func (ch SlackMessageChan) PostNopaste(np nopasteContent, url string) {
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
		IconURL:   np.IconURL,
		LinkNames: np.LinkNames,
	}
	select {
	case ch <- msg:
	default:
		log.Println("Can't send msg to Slack")
	}
}

func (ch SlackMessageChan) PostMsgr(req *http.Request) {
	username := req.FormValue("username")
	if username == "" {
		username = "msgr"
	}
	msg := SlackMessage{
		Channel:   req.FormValue("channel"),
		Text:      req.FormValue("msg"),
		Username:  username,
		IconEmoji: req.FormValue("icon_emoji"),
		IconURL:   req.FormValue("icon_url"),
		LinkNames: 0, // default notice as IRC
	}
	if _notice := req.FormValue("notice"); _notice == "0" {
		msg.LinkNames = 1
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

func (a *SlackAgent) Post(m SlackMessage) error {
	payload, _ := json.Marshal(&m)
	v := url.Values{}
	v.Set("payload", string(payload))
	log.Println("post to slack", a, string(payload))
	resp, err := a.client.PostForm(a.WebhookURL, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		return fmt.Errorf("failed post to slack:%s", body)
	} else {
		return err
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
	ignoreUntil := Epoch
	backoff := SlackInitialBackOff
	for {
		select {
		case msg := <-ch:
			if time.Now().Before(ignoreUntil) {
				// ignored
				continue
			}
			throttle(lastPostedAt, SlackThrottleWindow)
			err := agent.Post(msg)
			lastPostedAt = time.Now()
			if err != nil {
				backoff = int(math.Min(float64(backoff)*2, SlackMaxBackOff))
				d, _ := time.ParseDuration(fmt.Sprintf("%ds", backoff))
				ignoreUntil = lastPostedAt.Add(d)
				log.Println(err, msg.Channel, "will be ignored until", ignoreUntil)
			} else if !ignoreUntil.Equal(Epoch) {
				ignoreUntil = Epoch
				backoff = SlackInitialBackOff
			}
		}
	}
}
