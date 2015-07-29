package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	pit "github.com/typester/go-pit"
)

var (
	Endpoint string
)

func main() {
	// parse arguments
	var channel string
	var summary string
	var notice bool
	var useAuth bool
	var username string
	var password string

	flag.StringVar(&channel, "channel", "", "irc channel")
	flag.StringVar(&channel, "c", "", "irc channel")
	flag.StringVar(&summary, "summary", "", "text summary")
	flag.StringVar(&summary, "s", "", "text summary")
	flag.BoolVar(&notice, "notice", false, "send as notice")
	flag.BoolVar(&notice, "n", false, "send as notice")
	flag.BoolVar(&useAuth, "use_auth", false, "use auth from pit")
	flag.BoolVar(&useAuth, "u", false, "use auth from pit")
	flag.Parse()

	// get password if use_auth specified
	if useAuth {
		u, err := url.Parse(Endpoint)
		if err != nil {
			panic(err)
		}
		profile, err := pit.Get(u.Host, pit.Requires{"username": "username on nopaste", "password": "password on nopaste"})
		if err != nil {
			log.Fatal(err)
		}
		tmp_user, ok := (*profile)["username"]
		if !ok {
			log.Fatal("password is not found")
		}
		username = tmp_user

		tmp_pass, ok := (*profile)["password"]
		if !ok {
			log.Fatal("password is not found")
		}
		password = tmp_pass
	}

	bytes, err := ioutil.ReadAll(os.Stdin)

	values := make(url.Values)
	values.Set("text", string(bytes))
	if len(summary) > 0 {
		values.Set("summary", summary)
	}
	if len(channel) > 0 {
		values.Set("channel", "#"+channel)
	}
	if notice {
		values.Set("notice", "1")
	} else {
		values.Set("notice", "0")
	}

	request, err := http.NewRequest("POST", Endpoint, strings.NewReader(values.Encode()))
	request.ParseForm()
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if len(username) > 0 {
		request.SetBasicAuth(username, password)
	}

	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// goのhttpはデフォルトでリダイレクトを読みに行って
		// しかもAuth Headerを付けずに401で死ぬやつなのでここで
		// リダイレクトポリシーを設定して差し止める
		os.Stdout.WriteString("Nopaste URL: " + req.URL.String() + "\n")
		return errors.New("")
	}

	response, err := client.Do(request)
	if err != nil {
		// log.Fatal(err)
		return
	}
	defer response.Body.Close()
}
