package pocket

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

const (
	authorizeUrl = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"

	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"
)

type RequestTokenRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectURI string `json:"redirect_uri"`
}

type AuthorizeRequest struct {
	ConsumerKey string `json:"consumer_key"`
	Code        string `json:"code"`
}

type AuthorizeResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

// GetRequestToken obtains the request token that is used to authorize user in your application
func (c *Client) GetRequestToken(ctx context.Context, redirectUrl string) (string, error) {
	inp := &RequestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUrl,
	}

	values, err := c.DoHTTP(ctx, endpointRequestToken, inp)
	if err != nil {
		return "", err
	}

	if values.Get("code") == "" {
		return "", errors.New("empty request token in API response")
	}

	return values.Get("code"), nil
}

// GetAuthorizationURL generates link to authorize user
func (c *Client) GetAuthorizationURL(requestToken, redirectUrl string) (string, error) {
	if requestToken == "" || redirectUrl == "" {
		return "", errors.New("empty params")
	}

	return fmt.Sprintf(authorizeUrl, requestToken, redirectUrl), nil
}

// Authorize generates access token for user, that authorized in your app via link
func (c *Client) Authorize(ctx context.Context, requestToken string) (*AuthorizeResponse, error) {
	if requestToken == "" {
		return nil, errors.New("empty request token")
	}

	inp := &AuthorizeRequest{
		Code:        requestToken,
		ConsumerKey: c.consumerKey,
	}

	values, err := c.DoHTTP(ctx, endpointAuthorize, inp)
	if err != nil {
		return nil, err
	}

	accessToken, username := values.Get("access_token"), values.Get("username")
	if accessToken == "" {
		return nil, errors.New("empty access token in API response")
	}

	return &AuthorizeResponse{
		AccessToken: accessToken,
		Username:    username,
	}, nil
}
