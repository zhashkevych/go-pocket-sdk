package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	host         = "https://getpocket.com/v3"
	authorizeUrl = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"

	endpointAdd          = "/add"
	endpointModify       = "/modify"
	endpointRetrieve     = "/retrieve"
	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"

	xErrorHeader     = "X-Error"
	xErrorCodeHeader = "X-Error-Code"

	defaultTimeout = 5 * time.Second
)

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}

	requestTokenResponse struct {
		Code string `json:"code"`
	}

	authorizeRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	AuthorizeResponse struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	AddInput struct {
		URL         string
		Title       string
		Tags        []string
		AccessToken string
	}

	addRequest struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Tags        string `json:"tags"`
		AccessToken string `json:"access_token"`
		ConsumerKey string `json:"consumer_key"`
	}

	//errorResponse struct {
	//	Code    string
	//	Message string
	//}
)

func (i AddInput) generateRequest(consumerKey string) addRequest {
	return addRequest{
		URL:         url.QueryEscape(i.URL),
		Tags:        strings.Join(i.Tags, ","),
		Title:       i.Title,
		AccessToken: i.AccessToken,
		ConsumerKey: consumerKey,
	}
}

type Client struct {
	client      *http.Client
	consumerKey string
}

func NewClient(consumerKey string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		consumerKey: consumerKey,
	}
}

func (c *Client) GetRequestToken(ctx context.Context, redirectUrl string) (string, error) {
	resp := &requestTokenResponse{}
	inp := &requestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUrl,
	}

	if err := c.doHTTP(ctx, endpointRequestToken, inp, resp); err != nil {
		return "", err
	}

	return resp.Code, nil
}

func (c *Client) GetAuthorizationURL(requestToken, redirectUrl string) string {
	return fmt.Sprintf(authorizeUrl, requestToken, redirectUrl)
}

func (c *Client) Authorize(ctx context.Context, requestToken string) (*AuthorizeResponse, error) {
	resp := &AuthorizeResponse{}
	inp := &authorizeRequest{
		Code:        requestToken,
		ConsumerKey: c.consumerKey,
	}

	if err := c.doHTTP(ctx, endpointAuthorize, inp, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Add(ctx context.Context, input AddInput) error {
	req := input.generateRequest(c.consumerKey)

	if err := c.doHTTP(ctx, endpointAdd, req, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) Modify(ctx context.Context) {

}

func (c *Client) Retrieve(ctx context.Context) {

}
func (c *Client) doHTTP(ctx context.Context, endpoint string, body, response interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF8")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Header.Get(xErrorHeader))

	}

	if response == nil {
		return nil
	}

	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respB, response)
}
