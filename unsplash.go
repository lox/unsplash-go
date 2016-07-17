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

// Client is a client for Unsplash.com's API
type Client struct {
	Client        *http.Client
	requestFilter func(r *http.Request)
}

// NewClient creates a new Unsplash Client for unauthenticated public actions
func NewClient(clientID string) *Client {
	return &Client{
		Client: &http.Client{},
		requestFilter: func(r *http.Request) {
			r.Header.Set("Authorization", "Client-ID "+clientID)
		},
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

// GetUserPhotos gets all of a users photos
func (c *Client) GetUserPhotos(username string, orderBy string, f func(p Photo) (bool, error)) error {
	params := map[string]string{
		"order_by": orderBy,
	}

	return c.getPhotos(fmt.Sprintf("users/%s/photos", username), params, f)
}

// GetUsersLikes gets all the photos a user has liked
func (c *Client) GetUsersLikes(username string, orderBy string, f func(p Photo) (bool, error)) error {
	params := map[string]string{
		"order_by": orderBy,
	}

	return c.getPhotos(fmt.Sprintf("users/%s/likes", username), params, f)
}

// GetCollection gets all the photos from a collection
func (c *Client) GetCollection(id string, orderBy string, f func(p Photo) (bool, error)) error {
	params := map[string]string{
		"order_by": orderBy,
	}

	return c.getPhotos(fmt.Sprintf("collections/%s/photos", id), params, f)
}

func (c *Client) getPhotos(path string, params map[string]string, f func(p Photo) (bool, error)) error {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return err
	}

	for k, v := range params {
		req.URL.Query().Set(k, v)
	}

	return c.paginate(req, func(resp *http.Response) (bool, error) {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return true, err
		}

		resp.Body.Close()
		var photos []Photo

		if err = json.Unmarshal(b, &photos); err != nil {
			return true, err
		}

		for _, photo := range photos {
			if done, err := f(photo); done {
				return done, err
			}
		}

		return false, nil
	})
}

func (c *Client) paginate(req *http.Request, f func(resp *http.Response) (bool, error)) error {
	page := 1

	for {
		req.URL.Query().Set("page", strconv.Itoa(page))

		log.Printf("Requesting %s %s", req.Method, req.URL.String())
		resp, err := c.Client.Do(req)
		if err != nil {
			return err
		}

		log.Printf("Response was %d %s", resp.StatusCode, resp.Status)
		if done, err := f(resp); err != nil {
			return err
		} else if done {
			return nil
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

// Photo is an Unsplash Photo
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
