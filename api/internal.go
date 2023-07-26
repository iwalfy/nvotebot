package api

import (
	"net/url"
)

type baseResponse struct {
	Error_   bool   `json:"error"`
	Code_    int    `json:"code,omitempty"`
	Message_ string `json:"message,omitempty"`
}

type Response interface {
	Error() bool
	Code() int
	Message() string
}

func (r *baseResponse) Error() bool     { return r.Error_ }
func (r *baseResponse) Code() int       { return r.Code_ }
func (r *baseResponse) Message() string { return r.Message_ }

func createParams(params ...url.Values) url.Values {
	out := make(url.Values)
	for _, param := range params {
		for key, values := range param {
			for _, value := range values {
				out.Add(key, value)
			}
		}
	}
	return out
}
