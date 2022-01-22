package noclist

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	headerBadsecAuthenticationToken = "badsec-authentication-token"
	headerRequestChecksum           = "x-request-checksum"
)

var (
	ErrAuthentication = errors.New("client: authentication to server failed")
	ErrTimeout        = errors.New("client: hit maximum retries")
	ErrFailed         = errors.New("client: request failed")
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

func (c *Client) ListUsers() ([]string, error) {
	path := "/users"
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, ErrFailed // base URL was specified incorrectly
	}
	resp, err := c.doRetry(req)
	if err != nil || resp.StatusCode != 200 {
		return nil, ErrFailed
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrFailed
	}
	return strings.Split(string(b), "\n"), nil
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
