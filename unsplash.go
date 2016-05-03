package unsplash

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Client struct {
	client   http.Client
	clientID string
}

func NewClient(clientID string) *Client {
	return &Client{
		clientID: clientID,
	}
}

func (c *Client) GetUserPhotos(username string) ([]Photo, error) {
	req, err := http.NewRequest("GET", "https://api.unsplash.com/users/lox/photos", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Client-ID "+c.clientID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var photos []Photo

	if err = json.Unmarshal(b, &photos); err != nil {
		return nil, err
	}

	return photos, nil
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
