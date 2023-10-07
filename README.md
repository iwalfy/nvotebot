# nvotebot
Telegram Bot for Neural OpenNet (https://t.me/neuro_opennet) post voting system (https://nvote.lebedinets.ru/).

# Building:
```console
$ CGO_ENABLED=0 go build -o bot -v -tags osusergo,netgo -ldflags '-w -s' github.com/iwalfy/nvotebot/cmd/nvotebot
```

# Running:
```console
$ NVOTE_API_TOKEN=secret NVOTE_BOT_TOKEN=botfather ./bot
```
