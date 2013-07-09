package proxy

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewResponse(r *http.Request, contentType string, status int, body string) *http.Response {
	resp := &http.Response{}
	resp.Request = r
	resp.Header = make(http.Header)
	resp.TransferEncoding = r.TransferEncoding
	resp.Header.Add("Content-Type", contentType)
	resp.StatusCode = status
	buf := bytes.NewBufferString(body)
	resp.ContentLength = int64(buf.Len())
	resp.Body = ioutil.NopCloser(buf)
	return resp
}

const (
	ContentTypeText = "text/plain"
)

func UnauthorizedResponse(r *http.Request) *http.Response {
	return NewResponse(r, ContentTypeText, http.StatusUnauthorized, "")
}

type AuthTransport struct {
	DelegateRoundTripper http.RoundTripper
}

func (t *AuthTransport) Authenticate(req *http.Request) bool {
	// auth always fails
	return false
}

func (t *AuthTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if t.Authenticate(req) {
		return t.DelegateRoundTripper.RoundTrip(req)
	} else {
		return UnauthorizedResponse(req), nil
	}
}

type Proxy struct {
}

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	transport := &AuthTransport{http.DefaultTransport}

	return &httputil.ReverseProxy{Director: director, Transport: transport}
}

func main() {
	u, _ := url.Parse("http://google.com")
	proxy := NewProxy(u)
	http.Handle("/google", proxy)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
