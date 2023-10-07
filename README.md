# nvotebot
Telegram Bot for Neural OpenNet (https://t.me/neuro_opennet) post voting system (https://nvote.lebedinets.ru/).

# Running:
```console
$ nix run github:iwalfy/nvotebot
```

<details>
	<summary>Building</summary>

```console
$ CGO_ENABLED=0 go build -v -tags osusergo,netgo -ldflags '-w -s' github.com/iwalfy/nvotebot/cmd/nvotebot
```
</details>
