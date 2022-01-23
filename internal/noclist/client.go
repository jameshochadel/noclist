package noclist

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	headerBadsecAuthenticationToken = "badsec-authentication-token"
	headerRequestChecksum           = "x-request-checksum"
)

var (
	ErrAuthentication  = errors.New("authentication to server failed")
	ErrConfigServerURL = errors.New("config value 'ServerURL' was not a valid URL")
	ErrFailed          = errors.New("request failed")
	ErrInternal        = errors.New("an unexpected internal error occurred")
	ErrTimeout         = errors.New("hit maximum retries")
)

// Client wraps calls to the BADSEC API.
type Client struct {
	// baseURL is the scheme and host of the server, like "http://example.com".
	baseURL string
	client  http.Client
	token   string
}

func (c *Client) authenticate() error {
	req, err := http.NewRequest("HEAD", c.baseURL+"/auth", nil)
	if err != nil {
		return ErrInternal
	}
	resp, err := c.doRetry(req)
	if err != nil {
		return ErrAuthentication
	}
	c.token = resp.Header.Get(headerBadsecAuthenticationToken)
	return nil
}

// checksum should only be called on Clients created via noclist.New().
func (c *Client) checksum(path string) string {
	s := sha256.New()
	s.Write([]byte(c.token + path))
	return fmt.Sprintf("%x", s.Sum(nil))
}

// doRetry attempts to make a request up to three times. The request is
// retried if it fails at the network level or if the response has any status
// code besides 200. The x-request-checksum header is automatically added to
// the request.
func (c *Client) doRetry(req *http.Request) (*http.Response, error) {
	req.Header.Set(headerRequestChecksum, c.checksum(req.URL.Path))

	for i := 0; i < 3; i++ {
		resp, err := c.client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			return resp, nil
		}
	}

	return nil, ErrTimeout
}

func (c *Client) Authenticated() bool {
	return c.token != ""
}

// ListUsers requests the list of users from the server and returns a slice of
// user IDs if successful, or nil and an error if not.
func (c *Client) ListUsers() ([]string, error) {
	path := "/users"
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, ErrInternal
	}
	resp, err := c.doRetry(req)
	if err != nil || resp.StatusCode != 200 {
		return nil, ErrFailed
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrFailed
	}
	return strings.Split(strings.Trim(string(b), "\n"), "\n"), nil
}

// New returns a Client that is ready to make authenticated requests to the server,
// or an error if the Client could not be created.
//
// New makes a network request to the server to authenticate.
func New(cfg Config) (*Client, error) {
	if _, err := url.ParseRequestURI(cfg.ServerURL); err != nil {
		return nil, ErrConfigServerURL
	}
	c := Client{baseURL: cfg.ServerURL}
	err := c.authenticate()
	return &c, err
}
