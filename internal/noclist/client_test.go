package noclist_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jameshochadel/noclist/internal/noclist"
)

var (
	ResponseAuthSucceeded = serverResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{"badsec-authentication-token": []string{"something"}},
	}
	ResponseFailed = serverResponse{
		// any code != http.StatusOK is fine
		StatusCode: http.StatusForbidden,
	}
	ResponseListUsersSucceeded = serverResponse{
		StatusCode: http.StatusOK,
		Body:       "foo\nbar",
	}
)

// serverResponse consolidates the important fields of an HTTP response for
// use with the httptest server.
type serverResponse struct {
	StatusCode int
	Headers    http.Header
	Body       string
}

type testServer struct {
	Server       *httptest.Server
	RequestCount int
}

// NewServer configures a httptest.Server that responds to requests with
// each element in responses in order. If a request is made after every response
// in responses has been sent, the server will call t.Fail().
//
// The caller is responsible for calling s.Server.Close().
func NewServer(t *testing.T, responses []serverResponse) *testServer {
	t.Helper()

	s := testServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.RequestCount >= len(responses) {
			t.Log("more requests were mode than responses provided")
			t.Fail()
		}
		response := responses[s.RequestCount]
		for k, v := range response.Headers {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(response.StatusCode)
		_, err := w.Write([]byte(response.Body))
		if err != nil {
			t.Fatalf("httptest.Server: unexpected error writing response body")
		}
		s.RequestCount++
	}))
	return &s
}

func TestNew(t *testing.T) {
	cases := []struct {
		Name                string
		Responses           []serverResponse
		ClientAuthenticated bool
		Error               error
	}{
		{
			Name: "immediate success",
			Responses: []serverResponse{
				ResponseAuthSucceeded,
			},
			ClientAuthenticated: true,
			Error:               nil,
		},
		{
			Name: "repeated failure",
			Responses: []serverResponse{
				ResponseFailed,
				ResponseFailed,
				ResponseFailed,
			},
			ClientAuthenticated: false,
			Error:               noclist.ErrAuthentication,
		},
		{
			Name: "success after retries",
			Responses: []serverResponse{
				ResponseFailed,
				ResponseFailed,
				ResponseAuthSucceeded,
			},
			ClientAuthenticated: true,
			Error:               nil,
		},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf(tc.Name), func(t *testing.T) {
			// arrange
			s := NewServer(t, tc.Responses)
			defer s.Server.Close()

			// act
			client, err := noclist.New(noclist.Config{ServerURL: s.Server.URL})

			// assert
			if client.Authenticated() != tc.ClientAuthenticated {
				t.Logf("expected client created to be %v, got %v", tc.ClientAuthenticated, client.Authenticated())
				t.Fail()
			}
			if !errors.Is(err, tc.Error) {
				t.Logf("expected error to be %v, got %v", tc.Error, err)
				t.Fail()
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	cases := []struct {
		Name      string
		Responses []serverResponse
		Users     []string
		Error     error
	}{
		{
			Name: "immediate success",
			Responses: []serverResponse{
				ResponseAuthSucceeded,
				ResponseListUsersSucceeded,
			},
			Users: []string{"foo", "bar"},
			Error: nil,
		},
		{
			Name: "success after retries",
			Responses: []serverResponse{
				ResponseAuthSucceeded,
				ResponseFailed,
				ResponseFailed,
				ResponseListUsersSucceeded,
			},
			Users: []string{"foo", "bar"},
			Error: nil,
		},
		{
			Name: "repeated failure",
			Responses: []serverResponse{
				ResponseAuthSucceeded,
				ResponseFailed,
				ResponseFailed,
				ResponseFailed,
			},
			Users: nil,
			Error: noclist.ErrFailed,
		},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf(tc.Name), func(t *testing.T) {
			// arrange
			s := NewServer(t, tc.Responses)
			defer s.Server.Close()
			client, err := noclist.New(noclist.Config{ServerURL: s.Server.URL})
			if err != nil {
				t.Fatal("unexpected failure authenticating")
			}

			// act
			users, err := client.ListUsers()

			// assert
			if !reflect.DeepEqual(users, tc.Users) {
				t.Logf("expected users to be %v, got %v", tc.Users, users)
				t.Fail()
			}
			if !errors.Is(err, tc.Error) {
				t.Logf("expected error to be %v, got %v", tc.Error, err)
				t.Fail()
			}
		})
	}
}
