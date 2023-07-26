package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	. "github.com/iwalfy/nvotebot/util"
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
	Votes int `json:"votes,omitempty"`
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

func (c *Client) makeRequest(ctx context.Context, method string, params url.Values, out Response) error {
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

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("backend returned status != 200 (current: %d)", res.StatusCode)
	}

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
