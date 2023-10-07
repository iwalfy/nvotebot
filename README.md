# nvotebot
Telegram Bot for Neural OpenNet (https://t.me/neuro_opennet) post voting system (https://nvote.lebedinets.ru/).

# Building:
```console
$ nix build
```
<details>
	<summary>Or just...</summary>

```console
$ CGO_ENABLED=0 go build -v -tags osusergo,netgo -ldflags '-w -s' github.com/iwalfy/nvotebot/cmd/nvotebot
```
</details>

# Running:
```console
$ NVOTE_API_TOKEN=secret NVOTE_BOT_TOKEN=botfather ./result/bin/nvotebot
```

<details>
	<summary>Or just...</summary>
	
```console
$ ./nvotebot
```
</details>
