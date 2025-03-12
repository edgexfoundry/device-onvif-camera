package onvif

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DigestClient represents an HTTP client used for making requests authenticated
// with http digest authentication.
type DigestClient struct {
	client     *http.Client
	username   string
	password   string
	snonce     string
	realm      string
	qop        string
	nonceCount uint32
}

// NewDigestClient returns a DigestClient that wraps a given standard library http Client with the given username and password
func NewDigestClient(stdClient *http.Client, username string, password string) *DigestClient {
	return &DigestClient{
		client:   stdClient,
		username: username,
		password: password,
	}
}

func (dc *DigestClient) Do(httpMethod string, endpoint string, soap string) (*http.Response, error) {
	req, err := createHttpRequest(httpMethod, endpoint, soap)
	if err != nil {
		return nil, err
	}
	if dc.snonce != "" {
		digestAuth, err := dc.getDigestAuth(req.Method, req.URL.String())
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", digestAuth)
	}

	// Attempt the request using the underlying client
	resp, err := dc.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	dc.getDigestParts(resp)
	// We will need to return the response from another request, so defer a close on this one
	defer resp.Body.Close()

	req, err = createHttpRequest(httpMethod, endpoint, soap)
	if err != nil {
		return nil, err
	}
	digestAuth, err := dc.getDigestAuth(req.Method, req.URL.String())
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", digestAuth)

	authedResp, err := dc.client.Do(req)
	if err != nil {
		return nil, err
	}
	return authedResp, nil
}

func (dc *DigestClient) getDigestParts(resp *http.Response) {
	result := map[string]string{}
	authHeader := resp.Header.Get("WWW-Authenticate")
	if len(authHeader) > 0 {
		wantedHeaders := []string{"nonce", "realm", "qop"}
		responseHeaders := strings.Split(authHeader, ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	dc.snonce = result["nonce"]
	dc.realm = result["realm"]
	dc.qop = result["qop"]
	dc.nonceCount = 0
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() (string, error) {
	b := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", fmt.Errorf("generate random number failed: %w", err)
	}
	return fmt.Sprintf("%x", b)[:16], nil
}

func (dc *DigestClient) getDigestAuth(method string, uri string) (string, error) {
	ha1 := getMD5(dc.username + ":" + dc.realm + ":" + dc.password)
	ha2 := getMD5(method + ":" + uri)
	cnonce, err := getCnonce()
	if err != nil {
		return "", fmt.Errorf("get DigestAuth failed: %w", err)
	}
	dc.nonceCount++
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, dc.snonce, dc.nonceCount, cnonce, dc.qop, ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		dc.username, dc.realm, dc.snonce, uri, cnonce, dc.nonceCount, dc.qop, response)
	return authorization, nil
}
