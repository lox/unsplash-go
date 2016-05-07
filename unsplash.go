package unsplash

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/peterhellberg/link"
)

const (
	apiVersion = "v1"
)

type Client struct {
	client *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://api.unsplash.com/%s", path), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept-Version", apiVersion)
	return req, err
}

func (c *Client) GetUserPhotos(username string, f func(p Photo) error) error {
	req, err := c.newRequest("GET", fmt.Sprintf("users/%s/photos", username), nil)
	if err != nil {
		return err
	}

	return c.paginate(req, func(resp *http.Response) error {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		resp.Body.Close()
		var photos []Photo

		if err = json.Unmarshal(b, &photos); err != nil {
			return err
		}

		for _, photo := range photos {
			if err = f(photo); err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *Client) paginate(req *http.Request, f func(resp *http.Response) error) error {
	page := 1

	for {
		req.URL.Query().Set("page", strconv.Itoa(page))

		log.Printf("Requesting %s %s", req.Method, req.URL.String())
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		if err = f(resp); err != nil {
			return err
		}

		hasNext := false

		for _, l := range link.ParseResponse(resp) {
			if l.Rel == "next" {
				next, err := url.Parse(l.URI)
				if err != nil {
					return err
				}
				req.URL = next
				hasNext = true
			}
		}

		if !hasNext {
			break
		}
	}
	return nil
}

type Photo struct {
	ID          string `json:"id"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Color       string `json:"color"`
	Likes       int    `json:"likes"`
	LikedByUser bool   `json:"liked_by_user"`
	User        struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		Name         string `json:"name"`
		ProfileImage struct {
			Small  string `json:"small"`
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"profile_image"`
		Links struct {
			Self   string `json:"self"`
			HTML   string `json:"html"`
			Photos string `json:"photos"`
			Likes  string `json:"likes"`
		} `json:"links"`
	} `json:"user"`
	CurrentUserCollections []interface{} `json:"current_user_collections"`
	Urls                   struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
	Categories []interface{} `json:"categories"`
	Links      struct {
		Self     string `json:"self"`
		HTML     string `json:"html"`
		Download string `json:"download"`
	} `json:"links"`
}

type pager struct {
	lister func(page int) (v interface{}, nextPage int, err error)
}

func (p *pager) Pages(f func(v interface{}, lastPage bool) bool) error {
	page := 1
	for {
		val, nextPage, err := p.lister(page)
		if err != nil {
			return err
		}
		if !f(val, nextPage == 0) || nextPage == 0 {
			break
		}
		page = nextPage
	}
	return nil
}
