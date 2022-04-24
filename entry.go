package pocket

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

const (
	PocketDataURL = "https://getpocket.com/v3/get"
	Offset        = 50
)

type State string

const (
	UNREAD  State = "unread"
	ARCHIVE State = "archive"
	ALL     State = "all"
)

type Content string

const (
	ARTICLE Content = "content"
	VIDEO   Content = "video"
	IMAGE   Content = "image"
)

type Favorite uint

const (
	UNFAVORITED Favorite = 0
	FAVORITED   Favorite = 1
)

type Sort string

const (
	NEWEST Sort = "newest"
	OLDEST Sort = "oldest"
	TITLE  Sort = "title"
	SITE   Sort = "site"
)

type GetPocketDataBody struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`

	// unread = only return unread items (default)
	// archive = only return archived items
	// all = return both unread and archived items
	State State `json:"state,omitempty"`

	// 0 = only return un-favorited items
	// 1 = only return favorited items
	Favorite Favorite `json:"favorite,omitempty"`

	// tag_name = only return items tagged with tag_name
	// _untagged_ = only return untagged items
	Tag string `json:"tag,omitempty"`

	// article = only return articles
	// video = only return videos or articles with embedded videos
	// image = only return images
	ContentType Content `json:"contentType,omitempty"`

	// newest = return items in order of newest to oldest
	// oldest = return items in order of oldest to newest
	// title = return items in order of title alphabetically
	// site = return items in order of url alphabetically
	Sort Sort `json:"sort,omitempty"`

	// simple = return basic information about each item, including title, url, status, and more
	// complete = return all data about each item, including tags, images, authors, videos, and more
	DetailType string `json:"detailType,omitempty"`

	// Only return items whose title or url contain the search string
	Search string `json:"search,omitempty"`

	// Only return items from a particular domain
	Domain string `json:"domain,omitempty"`

	// Only return items modified since the given since unix timestamp
	Since int64 `json:"since,omitempty"`

	// Only return count number of items
	Count int `json:"count,omitempty"`

	// Used only with count; start returning from offset position of results
	Offset int `json:"offset,omitempty"`
}

// GetUnreadEntries
func (c *Client) GetUnreadEntries(accessToken string) error {
	// ctx := context.Background()
	// getPocketDataBody := GetPocketDataBody{
	// 	AccessToken: accessToken,
	// 	State:       UNREAD,
	// 	Sort:        OLDEST,
	// 	Offset:      Offset,
	// }

	// jsonBodyData, err := json.Marshal(getPocketDataBody)
	// if err != nil {
	// 	return fmt.Errorf("error while marshalling 'get unread entries body': %v", err)
	// }

	// req, err := http.NewRequest(http.MethodGet, PocketDataURL, bytes.NewBuffer(jsonBodyData))
	// if err != nil {
	// 	return fmt.Errorf("error while creating new 'get unread entries' %v", err)
	// }

	// respBody, err := c.DoHTTP(ctx, "", nil)
	// if err != nil {
	// 	return fmt.Errorf("error while doing the request: %v", err)
	// }

	// fmt.Println(string(respBody))
	return nil
}

type addRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Tags        string `json:"tags,omitempty"`
	AccessToken string `json:"access_token"`
	ConsumerKey string `json:"consumer_key"`
}

// AddInput holds data necessary to create new item in Pocket list
type AddInput struct {
	URL         string
	Title       string
	Tags        []string
	AccessToken string
}

func (i AddInput) validate() error {
	if i.URL == "" {
		return errors.New("required URL values is empty")
	}

	if i.AccessToken == "" {
		return errors.New("access token is empty")
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

// Add creates new item in Pocket list
func (c *Client) Add(ctx context.Context, input AddInput) error {
	if err := input.validate(); err != nil {
		return err
	}

	req := input.generateRequest(c.consumerKey)
	_, err := c.DoHTTP(ctx, endpointAdd, req)

	return err
}
