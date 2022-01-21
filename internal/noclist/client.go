package noclist

import (
	"errors"
	"net/http"
)

const (
	headerBadsecAuthenticationToken = "badsec-authentication-token"
	headerRequestChecksum           = "x-request-checksum"
)

var (
	ErrAuthentication = errors.New("client: authentication to server failed")
	ErrTimeout        = errors.New("client: hit maximum retries")
)

// Client wraps calls to the BADSEC API.
type Client struct {
	// baseURL is the scheme and host of the server, like "http://example.com".
	baseURL string
	client  http.Client
	token   string
}

func (c *Client) authenticate() error {
	resp, err := c.client.Head(c.baseURL + "/auth")
	if err != nil || resp.StatusCode != 200 {
		return ErrAuthentication
	}
	c.token = resp.Header.Get(headerBadsecAuthenticationToken)
	return nil
}

// New returns a Client that is ready to make authenticated requests to the server,
// or an error if the Client could not be created.
//
// New makes a network request to the server to authenticate.
func New(cfg Config) (*Client, error) {
	c := Client{baseURL: cfg.ServerURL}
	err := c.authenticate()
	return &c, err
}
