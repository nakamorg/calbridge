package http

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
)

const (
	authHeader = "WWW-Authenticate"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type digestAuthHTTPClient struct {
	c                  HTTPClient
	username, password string
}

type basicAuthHTTPClient struct {
	c                  HTTPClient
	username, password string
}

// HTTPClientWithBasicAuth returns an HTTP client that adds basic
// authentication to all outgoing requests. If c is nil, http.DefaultClient is
// used.
func HTTPClientWithBasicAuth(c HTTPClient, username, password string) HTTPClient {
	if c == nil {
		c = http.DefaultClient
	}
	return &basicAuthHTTPClient{c, username, password}
}

func (c *basicAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.c.Do(req)
}

// HTTPClientWithDigestAuth returns an HTTP client that adds basic
// authentication to all outgoing requests. If c is nil, http.DefaultClient is
// used.
func HTTPClientWithDigestAuth(c HTTPClient, username, password string) HTTPClient {
	if c == nil {
		c = http.DefaultClient
	}
	return &digestAuthHTTPClient{c, username, password}
}

func (c *digestAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// We need to copy the request body
	reqBody, err := req.GetBody()
	if err != err {
		return nil, err
	}
	resp, err := c.c.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		challenge := resp.Header.Get(authHeader)
		if len(challenge) == 0 {
			return resp, fmt.Errorf("empty challenge header")
		}
		digestAuthorization := digestHeader(c.username, c.password, req.Method, req.URL.String(), challenge)
		// Create a new request with the same URL and body as the original request
		newReq := req.Clone(context.Background())
		newReq.Body = reqBody
		newReq.Header.Set("Authorization", digestAuthorization)
		return http.DefaultClient.Do(newReq)
	}
	return nil, fmt.Errorf("server did not support digest auth. Response code: %v", resp.StatusCode)
}

func digestHeader(username, password, method, uri, challenge string) string {
	fields := parseDigestChallenge(challenge)

	realm := fields["realm"]
	nonce := fields["nonce"]
	qop := fields["qop"]
	opaque := fields["opaque"]
	algorithm := fields["algorithm"]

	// Using fixed cnonce. We should generate some random nonce though
	cnonce := "0a4f113b"
	// Generate nc (nonce count)
	nc := "00000001"

	ha1 := getMD5(username + ":" + realm + ":" + password)
	ha2 := getMD5(method + ":" + uri)
	response := getMD5(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":" + qop + ":" + ha2)

	// Construct the Authorization header
	authorization := fmt.Sprintf("Digest username=%q,realm=%q,nonce=%q,uri=%q,cnonce=%q,nc=%q,qop=%q,response=%q,opaque=%q,algorithm=%q",
		username, realm, nonce, uri, cnonce, nc, qop, response, opaque, algorithm)
	return authorization
}

func parseDigestChallenge(challenge string) map[string]string {
	result := make(map[string]string)

	// Remove the "Digest " prefix
	challenge = strings.TrimPrefix(challenge, "Digest ")

	// Split the challenge into key-value pairs
	pairs := strings.Split(challenge, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(parts[1], "\"")
			result[key] = value
		}
	}
	return result
}

func getMD5(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}
