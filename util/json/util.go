package json

import (
	"github.com/goccy/go-json"
	. "github.com/iwalfy/nvotebot/util"
)

func Stringify(v any) string {
	return string(Must(json.Marshal(v)))
}
