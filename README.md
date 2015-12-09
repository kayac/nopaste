nopaste
=============

nopaste http server & IRC agent.

Install
------

```
$ go get github.com/kayac/nopaste/cmd/nopaste
```

Usage
------

```
$ nopaste -config config.yaml
```

nopaste will rewrite the `config.yaml`(irc.channels section) when join to a new IRC channel.

Configuration
------

```yaml
base_url: http://example.com  # for IRC messages
listen: "localhost:3000"
data_dir: data
irc:
  host: localhost
  port: 6666
  secure: false
  password: secret
  nick: npbot
slack:
  webhook_url: https://hooks.slack.com/services/XXX/YYY/zzzz
channels:
- '#general'
- '#infra'
```

nopaste runs http server on `http://#{listen}/np`.

Amazon SNS http endpoint
-------

nopaste can accept Amazon SNS notification messages.

Set environment variables.

- `AWS_ACCESS_KEY_ID=XXXXXXXXXX`
- `AWS_SECRET_ACCESS_KEY=yyyyyyyyyyy`

Add a SNS topic http(s) endpoint to `http://example.com/np/amazon-sns/{channel_name}?key={message key}`.

LICENCE
-------

The MIT License (MIT)
