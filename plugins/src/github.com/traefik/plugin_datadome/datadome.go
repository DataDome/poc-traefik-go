package plugin_datadome

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Endpoint string `json:"endpoint"`
}

func CreateConfig() *Config {
	return &Config{}
}

type DataDomePlugin struct {
	next     http.Handler
	name     string
	endpoint string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &DataDomePlugin{
		next:     next,
		name:     name,
		endpoint: config.Endpoint,
	}, nil
}

func (m *DataDomePlugin) getClientId(req *http.Request) string {
	clientIDHeaders := req.Header.Get("x-datadome-clientid")
	if len(clientIDHeaders) > 0 {
		return clientIDHeaders
	}

	cookie, err := req.Cookie("datadome")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func (m *DataDomePlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Example of payload, values should come from the req
	DATADOME_SERVER_SIDE_KEY := os.Getenv("DATADOME_SERVER_SIDE_KEY")
	payload := url.Values{}
	payload.Set("User-Agent", req.Header.Get("User-Agent"))
	payload.Set("X-Forwarded-For", req.Header.Get("X-Forwarded-For"))
	payload.Set("Key", DATADOME_SERVER_SIDE_KEY)
	payload.Set("RequestModuleName", "traefik")
	payload.Set("ModuleVersion", "1.0.0")
	payload.Set("ServerName", "example-server")
	payload.Set("IP", "192.168.1.1")
	payload.Set("Protocol", "HTTP/1.1")
	payload.Set("Method", "GET")
	payload.Set("ServerHostName", "example.com")
	payload.Set("Request", "/api/protect")
	payload.Set("HeadersList", "Content-Type: application/json; Accept: */*")
	payload.Set("Host", req.Host)
	payload.Set("UserAgent", req.Header.Get("User-Agent"))
	payload.Set("XForwardedForIP", "203.0.113.1")
	payload.Set("XRealIP", "198.51.100.1")
	payload.Set("ClientID", m.getClientId(req))

	// Calling DataDome API to validate the request
	resp, err := http.Post(
		"http://api.datadome.co/validate-request",
		"application/x-www-form-urlencoded",
		strings.NewReader(payload.Encode()),
	)

	if err != nil || resp == nil {
		fmt.Println("Error: calling the API", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Internal error"))
		return //we should fail open, just an example here
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	// if err != null...

	// Handle response headers
	// https://docs.datadome.co/reference/validate-request#handling-api-response-headers
	cookie := resp.Header.Get("Set-Cookie")

	if resp.StatusCode == http.StatusForbidden {
		rw.Header().Add("X-test", "Blocked")
		rw.Header().Add("Set-Cookie", cookie)
		rw.WriteHeader(http.StatusForbidden)
		// Presenting the body of the response to the client (Challenge/Captcha)
		rw.Write([]byte(bodyBytes))
		return
	} else if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: unexpected response", resp.StatusCode, string(bodyBytes))
		rw.WriteHeader(http.StatusInternalServerError)
		return //we should fail open, just an example here
	}

	// We need to inject the DataDome headers into the response headers coming back to the client after hit the origin server
	wrapped := &responseHeaderWrapper{
		ResponseWriter:  rw,
		headersToInject: map[string]string{"X-test": "Passed", "Set-Cookie": cookie},
	}

	m.next.ServeHTTP(wrapped, req)
}

type responseHeaderWrapper struct {
	http.ResponseWriter
	headersToInject map[string]string
	wroteHeader     bool
}

func (w *responseHeaderWrapper) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		for k, v := range w.headersToInject {
			w.Header().Set(k, v)
		}
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
