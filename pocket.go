package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	host         = "https://getpocket.com/v3"
	authorizeUrl = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"

	endpointAdd          = "/add"
	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"

	// xErrorHeader used to parse error message from Headers on non-2XX responses
	xErrorHeader = "X-Error"

	defaultTimeout = 5 * time.Second
)

var (
	ErrURLValueIsEmpty             = errors.New("required URL value is empty")
	ErrAccessTokenIsEmpty          = errors.New("access token is empty")
	ErrConsumerKeyIsEmpty          = errors.New("consumer key is empty")
	ErrEmptyRequestTokenInResponse = errors.New("empty request token in API response")
	ErrEmptyParams                 = errors.New("empty params")
	ErrEmptyRequestToken           = errors.New("empty request token")
	ErrEmptyAccessTokenInResponse  = errors.New("empty access token in API response")
)

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}

	authorizeRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	AuthorizeResponse struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	addRequest struct {
		URL         string `json:"url"`
		Title       string `json:"title,omitempty"`
		Tags        string `json:"tags,omitempty"`
		AccessToken string `json:"access_token"`
		ConsumerKey string `json:"consumer_key"`
	}

	// AddInput holds data necessary to create new item in Pocket list
	AddInput struct {
		URL         string
		Title       string
		Tags        []string
		AccessToken string
	}

	// ClientOption allows manipulating client settings
	ClientOption func(*Client)
)

func (i AddInput) validate() error {
	if i.URL == "" {
		return ErrURLValueIsEmpty
	}

	if i.AccessToken == "" {
		return ErrAccessTokenIsEmpty
	}

	return nil
}

func (i AddInput) generateRequest(consumerKey string) addRequest {
	return addRequest{
		URL:         i.URL,
		Tags:        strings.Join(i.Tags, ","),
		Title:       i.Title,
		AccessToken: i.AccessToken,
		ConsumerKey: consumerKey,
	}
}

// Client is a getpocket API client
type Client struct {
	client      *http.Client
	consumerKey string
}

// NewClient creates a new client instance with your app key (to generate key visit https://getpocket.com/developer/apps/)
func NewClient(consumerKey string, options ...ClientOption) (*Client, error) {
	if consumerKey == "" {
		return nil, ErrConsumerKeyIsEmpty
	}

	client := &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		consumerKey: consumerKey,
	}

	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

// GetRequestToken obtains the request token that is used to authorize user in your application
func (c *Client) GetRequestToken(ctx context.Context, redirectUrl string) (string, error) {
	inp := &requestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUrl,
	}

	values, err := c.doHTTP(ctx, endpointRequestToken, inp)
	if err != nil {
		return "", err
	}

	if values.Get("code") == "" {
		return "", ErrEmptyRequestTokenInResponse
	}

	return values.Get("code"), nil
}

// GetAuthorizationURL generates link to authorize user
func (c *Client) GetAuthorizationURL(requestToken, redirectUrl string) (string, error) {
	if requestToken == "" || redirectUrl == "" {
		return "", ErrEmptyParams
	}

	return fmt.Sprintf(authorizeUrl, requestToken, redirectUrl), nil
}

// Authorize generates access token for user, that authorized in your app via link
func (c *Client) Authorize(ctx context.Context, requestToken string) (*AuthorizeResponse, error) {
	if requestToken == "" {
		return nil, ErrEmptyRequestToken
	}

	inp := &authorizeRequest{
		Code:        requestToken,
		ConsumerKey: c.consumerKey,
	}

	values, err := c.doHTTP(ctx, endpointAuthorize, inp)
	if err != nil {
		return nil, err
	}

	accessToken, username := values.Get("access_token"), values.Get("username")
	if accessToken == "" {
		return nil, ErrEmptyAccessTokenInResponse
	}

	return &AuthorizeResponse{
		AccessToken: accessToken,
		Username:    username,
	}, nil
}

// Add creates new item in Pocket list
func (c *Client) Add(ctx context.Context, input AddInput) error {
	if err := input.validate(); err != nil {
		return err
	}

	req := input.generateRequest(c.consumerKey)
	_, err := c.doHTTP(ctx, endpointAdd, req)

	return err
}

func (c *Client) doHTTP(ctx context.Context, endpoint string, body interface{}) (url.Values, error) {
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

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.client = httpClient
	}
}
