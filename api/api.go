package api

import (
	"context"
	"github.com/goccy/go-json"
	. "github.com/iwalfy/nvotebot/util"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	methodGet  = "get"
	methodVote = "vote"
)

type Client struct {
	apiURL string
	apiKey string
	client *http.Client
}

func NewClient(apiUrl, apiKey string) *Client {
	return &Client{
		apiURL: apiUrl,
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type baseResponse struct {
	Error_   bool   `json:"error"`
	Code_    int    `json:"code,omitempty"`
	Message_ string `json:"message,omitempty"`
}

func (r *baseResponse) Error() bool { return r.Error_ }

func (r *baseResponse) Code() int { return r.Code_ }

func (r *baseResponse) Message() string { return r.Message_ }

type IBaseResponse interface {
	Error() bool
	Code() int
	Message() string
}

type GetResponse struct {
	baseResponse
	ToVote []struct {
		UUID  string `json:"uuid"`
		Text  string `json:"text"`
		Votes int    `json:"votes"`
	} `json:"tovote,omitempty"`
}

type VoteResponse struct {
	baseResponse
	Votes int `json:"votes"`
}

type Error struct {
	message string
}

func (err *Error) Error() string {
	return err.message
}

func (c *Client) Get(ctx context.Context, userId int64, limit int) (*GetResponse, error) {
	res := &GetResponse{}
	err := c.makeRequest(ctx, methodGet, url.Values{
		"user_id": {strconv.FormatInt(userId, 10)},
		"limit":   {strconv.Itoa(limit)},
	}, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) Vote(ctx context.Context, userId int64, uuid string) (*VoteResponse, error) {
	res := &VoteResponse{}
	err := c.makeRequest(ctx, methodVote, url.Values{
		"user_id": {strconv.FormatInt(userId, 10)},
		"uuid":    {uuid},
	}, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) makeRequest(ctx context.Context, method string, params url.Values, out IBaseResponse) error {
	req := Must(http.NewRequestWithContext(
		ctx,
		"GET",
		c.apiURL,
		nil,
	))

	req.URL.RawQuery = createParams(params, url.Values{
		method:  {""},
		"token": {c.apiKey},
	}).Encode()

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, out)
	if err != nil {
		return err
	}

	if !out.Error() {
		return nil
	} else {
		return &Error{message: out.Message()}
	}
}

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
