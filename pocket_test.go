package pocket

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func newClient(t *testing.T, statusCode int, path, body string) *Client {
	return &Client{
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, path, r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		},
		consumerKey: "key",
	}
}

func TestClient_GetAuthorizationURL(t *testing.T) {
	type args struct {
		requestToken string
		redirectUrl  string
	}

	want := func(args args) string {
		return fmt.Sprintf("https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s", args.requestToken, args.redirectUrl)
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				requestToken: "qwe-rty-123",
				redirectUrl:  "http://localhost:80/",
			},
			wantErr: false,
		},
		{
			name: "Empty token",
			args: args{
				requestToken: "",
				redirectUrl:  "http://localhost:80/",
			},
			wantErr: true,
		},
		{
			name: "Empty URL",
			args: args{
				requestToken: "qwe-rty-123",
				redirectUrl:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got, err := c.GetAuthorizationURL(tt.args.requestToken, tt.args.redirectUrl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, want(tt.args), got)
			}
		})
	}
}

func TestClient_GetRequestToken(t *testing.T) {
	type args struct {
		ctx         context.Context
		redirectUrl string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedStatusCode   int
		expectedResponse     string
		expectedErrorMessage string
		want                 string
		wantErr              bool
	}{
		{
			name: "Ok",
			args: args{
				ctx:         context.Background(),
				redirectUrl: "http://localhost",
			},
			expectedStatusCode: 200,
			expectedResponse:   "code=qwe-rty-123",
			want:               "qwe-rty-123",
			wantErr:            false,
		},
		{
			name: "Empty redirect URL",
			args: args{
				ctx:         context.Background(),
				redirectUrl: "",
			},
			wantErr: true,
		},
		{
			name: "Empty response code",
			args: args{
				ctx:         context.Background(),
				redirectUrl: "http://localhost",
			},
			expectedStatusCode: 200,
			expectedResponse:   "code=",
			wantErr:            true,
		},
		{
			name: "Non-2XX Response",
			args: args{
				ctx:         context.Background(),
				redirectUrl: "http://localhost",
			},
			expectedStatusCode: 400,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.expectedStatusCode, "/v3/oauth/request", tt.expectedResponse)

			got, err := c.GetRequestToken(tt.args.ctx, tt.args.redirectUrl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_Authorize(t *testing.T) {

	type args struct {
		ctx          context.Context
		requestToken string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedStatusCode   int
		expectedResponse     string
		expectedErrorMessage string
		want                 *AuthorizeResponse
		wantErr              bool
	}{
		{
			name: "Ok",
			args: args{
				ctx:          context.Background(),
				requestToken: "token",
			},
			expectedResponse:   "access_token=qwe-rty-123&username=testuser",
			expectedStatusCode: 200,
			want: &AuthorizeResponse{
				AccessToken: "qwe-rty-123",
				Username:    "testuser",
			},
		},
		{
			name: "Empty token",
			args: args{
				ctx:          context.Background(),
				requestToken: "",
			},
			wantErr: true,
		},
		{
			name: "Empty Access Token in response",
			args: args{
				ctx:          context.Background(),
				requestToken: "token",
			},
			expectedResponse:   "username=testuser",
			expectedStatusCode: 200,
			wantErr:            true,
		},
		{
			name: "Non-2XX response",
			args: args{
				ctx:          context.Background(),
				requestToken: "token",
			},
			expectedStatusCode: 400,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.expectedStatusCode, "/v3/oauth/authorize", tt.expectedResponse)

			got, err := c.Authorize(tt.args.ctx, tt.args.requestToken)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_Add(t *testing.T) {
	type args struct {
		ctx   context.Context
		input AddInput
	}
	tests := []struct {
		name               string
		args               args
		expectedStatusCode int
		wantErr            bool
	}{
		{
			name: "Ok",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					URL: "http://example.link",
					AccessToken: "token",
				},
			},
			expectedStatusCode: 200,
		},
		{
			name: "Empty URL",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					AccessToken: "token",
				},
			},
			wantErr: true,
		},
		{
			name: "Empty Token",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					URL: "http://example.link",
				},
			},
			wantErr: true,
		},
		{
			name: "With Title",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					URL: "http://example.link",
					AccessToken: "token",
					Title: "example",
				},
			},
			expectedStatusCode: 200,
		},
		{
			name: "With Tags",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					URL: "http://example.link",
					AccessToken: "token",
					Title: "example",
					Tags: []string{"qwe", "rty", "123"},
				},
			},
			expectedStatusCode: 200,
		},
		{
			name: "Non-2XX Response",
			args: args{
				ctx: context.Background(),
				input: AddInput{
					URL: "http://example.link",
					AccessToken: "token",
					Title: "example",
					Tags: []string{"qwe", "rty", "123"},
				},
			},
			expectedStatusCode: 400,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.expectedStatusCode, "/v3/add", "")

			if err := c.Add(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
