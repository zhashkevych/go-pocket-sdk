package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	host = "https://getpocket.com/v3"

	endpointAdd = "/add"

	// xErrorHeader used to parse error message from Headers on non-2XX responses
	xErrorHeader = "X-Error"

	defaultTimeout = 5 * time.Second
)

// Client is a getpocket API client
type Client struct {
	client      *http.Client
	consumerKey string
}

// NewClient creates a new client instance with your app key (to generate key visit https://getpocket.com/developer/apps/)
func NewClient(consumerKey string) (*Client, error) {
	if consumerKey == "" {
		return nil, errors.New("consumer key is empty")
	}

	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		consumerKey: consumerKey,
	}, nil
}

func (c *Client) DoHTTP(ctx context.Context, endpoint string, body interface{}) (url.Values, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to marshal input body")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to create new request")
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF8")

	resp, err := c.client.Do(req)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to send http request")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("API Error: %s", resp.Header.Get(xErrorHeader))
		return url.Values{}, errors.New(err)
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to read request body")
	}

	values, err := url.ParseQuery(string(respB))
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to parse response body")
	}

	return values, nil
}
