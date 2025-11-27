package modulego

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func setupRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Host = "www.example.com"
	request.RemoteAddr = "127.0.0.1:1234"
	request.Proto = "http"
	request.Method = "GET"
	request.Header.Set("Hello", "world")
	request.Header.Set("user-agent", "über cool mozilla")
	request.Header.Set("referer", "www.example2.com")
	request.Header.Set("accept", "application/json")
	request.Header.Set("accept-encoding", "fr-FR")
	request.Header.Set("accept-charset", "utf8")
	request.Header.Set("origin", "www.example.com")
	request.Header.Set("x-forwarded-for", "192.168.10.10, 127.0.0.1")
	request.Header.Set("x-requested-with", "über_script")
	request.Header.Set("connection", "new")
	request.Header.Set("pragma", "no-cache")
	request.Header.Set("cache-control", "max-age=604800")
	request.Header.Set("x-real-ip", "127.0.0.1")

	return request
}

func DoCall(t *testing.T, url string, method string) (*http.Response, http.ResponseWriter, error) {
	client := &http.Client{}
	request, _ := http.NewRequest(method, url, nil)
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("error when making request to %s: %w", url, err)
	}

	rw := httptest.NewRecorder()
	_, err = io.Copy(rw, response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error when copying the response body: %w", err)
	}
	return response, rw, err
}

func TestNewClient(t *testing.T) {
	t.Run("Instantiate client with default values", func(t *testing.T) {
		serverSideKey := "your-api-key"
		c, err := NewClient(serverSideKey)

		assert.Nil(t, err)
		assert.Equal(t, DefaultEnableGraphQLSupportValue, c.EnableGraphQLSupport)
		assert.Equal(t, DefaultEnableReferrerRestorationValue, c.EnableReferrerRestoration)
		assert.Equal(t, DefaultEndpointValue, c.Endpoint)
		assert.NotNil(t, c.Logger)
		assert.Equal(t, DefaultMaximumBodySizeValue, c.MaximumBodySize)
		assert.Equal(t, DefaultModuleNameValue, c.ModuleName)
		assert.Equal(t, DefaultModuleVersionValue, c.ModuleVersion)
		assert.Equal(t, serverSideKey, c.ServerSideKey)
		assert.Equal(t, DefaultTimeoutValue, c.Timeout)
		assert.Equal(t, DefaultUrlPatternInclusionValue, c.UrlPatternInclusion)
		assert.Equal(t, DefaultUrlPatternExclusionValue, c.UrlPatternExclusion)
		assert.Equal(t, DefaultUseXForwardedHostValue, c.UseXForwardedHost)

		assert.NotNil(t, c.httpClient)
		assert.NotNil(t, c.urlPatternExclusion)
		assert.Nil(t, c.urlPatternInclusion)
	})

	t.Run("Error is returned when passing an empty string for the server-side key", func(t *testing.T) {
		c, err := NewClient("")

		assert.Nil(t, c)
		assert.NotNil(t, err)
		assert.Equal(t, "ServerSideKey must be defined", err.Error())
	})
}

func TestBuildRequest(t *testing.T) {
	dd, err := NewClient("Ob1w4n K3n0by")
	assert.Nil(t, err)

	request := setupRequest()
	result, err := dd.buildRequest(request)

	assert.Equal(t, nil, err)
	expectedResult := fmt.Sprintf("APIConnectionState=new&Accept=application%%2Fjson&AcceptCharset=utf8&AcceptEncoding=fr-FR&AuthorizationLen=0&CacheControl=max-age%%3D604800&Connection=new&CookiesLen=0&HeadersList=Accept-Encoding%%2COrigin%%2CX-Requested-With%%2CHello%%2CUser-Agent%%2CReferer%%2CAccept%%2CCache-Control%%2CX-Real-Ip%%2CAccept-Charset%%2CX-Forwarded-For%%2CConnection%%2CPragma&Host=www.example.com&IP=127.0.0.1&Key=Ob1w4n+K3n0by&Method=GET&ModuleVersion=%s&Origin=www.example.com&PostParamLen=0&Pragma=no-cache&Protocol=http&Referer=www.example2.com&Request=%%2Fping&RequestModuleName=%s&ServerHostname=www.example.com&Port=80&ServerName=www.example.com&TimeRequest=1695386441016659&UserAgent=%%C3%%BCber+cool+mozilla&X-Real-IP=127.0.0.1&X-Requested-With=%%C3%%BCber_script&XForwardedForIP=192.168.10.10%%2C+127.0.0.1", DefaultModuleVersionValue, DefaultModuleNameValue)

	if len(expectedResult) != len(result) {
		t.Error("Result length don't match")
		return
	}
}

func TestAddDataDomeHeaders(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock http call
	httpmock.RegisterResponder("GET", "/ping",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "pong")
			resp.Header.Add("X-Result", "true")
			return resp, nil
		},
	)

	// Mock API server call
	httpmock.RegisterResponder("POST", "/validate-request",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"key": "value",
			})
			resp.Header.Add("X-Datadome-Headers", "X-Datadome Set-Cookie")
			resp.Header.Add("X-Datadome", "protected")
			resp.Header.Add("Set-Cookie", "datadome=other_value")
			return resp, err
		},
	)

	_, origResp, err := DoCall(t, "/ping", http.MethodGet)
	assert.Equal(t, nil, err)
	origResp.Header().Add("Set-Cookie", "some_cookie=some_value")
	ddResp, _, err := DoCall(t, "/validate-request", http.MethodPost)
	assert.Equal(t, nil, err)

	origResp = addDataDomeHeaders(ddResp, origResp)

	assert.Equal(t, "", origResp.Header().Get("X-Datadome-Headers"))
	assert.Equal(t, "protected", origResp.Header().Get("X-Datadome"))
	for _, cookie := range origResp.Header()["Set-Cookie"] {
		assert.Contains(t, []string{"datadome=other_value", "some_cookie=some_value"}, cookie)
	}
}

func TestDefaultEndpointOnCustomDomain(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock http call
	httpmock.RegisterResponder("GET", "/ping",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "pong")
			resp.Header.Add("X-Result", "true")
			return resp, nil
		},
	)

	// Mock API server call
	httpmock.RegisterResponder("POST", "/validate-request",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"key": "value",
			})
			resp.Header.Add("X-Datadome-Headers", "X-Datadome")
			resp.Header.Add("X-Datadome", "protected")
			resp.Header.Add("X-Datadomeresponse", "200")
			return resp, err
		},
	)

	ddStruct, err := NewClient("azerty")
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	isBlocked, err := ddStruct.DatadomeProtect(rw, r)
	assert.Empty(t, err)
	assert.False(t, isBlocked)
}

func TestDatadomeProtect_NotBlocked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock http call
	httpmock.RegisterResponder("GET", "/ping",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "pong")
			resp.Header.Add("X-Result", "true")
			return resp, nil
		},
	)

	// Mock API server call
	httpmock.RegisterResponder("POST", "/validate-request",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"key": "value",
			})
			resp.Header.Add("X-Datadome-Headers", "X-Datadome")
			resp.Header.Add("X-Datadome", "protected")
			resp.Header.Add("X-Datadomeresponse", "200")
			return resp, err
		},
	)

	ddStruct, err := NewClient("azerty")
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	isBlocked, err := ddStruct.DatadomeProtect(rw, r)
	assert.Empty(t, err)
	assert.False(t, isBlocked)
}

func TestDatadomeProtect_Blocked(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock http call
	httpmock.RegisterResponder("GET", "/ping",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "pong")
			resp.Header.Add("X-Result", "true")
			return resp, nil
		},
	)

	// Mock API server call
	httpmock.RegisterResponder("POST", "/validate-request",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(403, map[string]interface{}{
				"key": "value",
			})
			resp.Header.Add("X-Datadome-Headers", "X-Datadome")
			resp.Header.Add("X-Datadome", "protected")
			resp.Header.Add("X-Datadomeresponse", "403")
			return resp, err
		},
	)

	ddStruct, err := NewClient("azerty")
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	isBlocked, err := ddStruct.DatadomeProtect(rw, r)
	assert.Empty(t, err)
	assert.True(t, isBlocked)
}

func TestDatadomeProtect_ExcludedPath(t *testing.T) {
	client, err := NewClient("azerty")
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/picture.jpg", nil)
	isBlocked, err := client.DatadomeProtect(rw, r)
	assert.Empty(t, err)
	assert.False(t, isBlocked)
}

func TestDatadomeProtect_IncludedPath(t *testing.T) {
	client, err := NewClient("azerty", WithUrlPatternInclusion(`(?i).*/included-path/?.*`))
	assert.Nil(t, err)

	rw := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/not-included-path", nil)
	isBlocked, err := client.DatadomeProtect(rw, r)
	assert.Empty(t, err)
	assert.False(t, isBlocked)
}

func TestAddDataDomeRequestHeaders(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock API server call
	httpmock.RegisterResponder("POST", "/validate-request",
		func(req *http.Request) (*http.Response, error) {
			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"key": "value",
			})
			resp.Header.Add("X-Datadome-Request-Headers", "X-Datadome-isbot")
			resp.Header.Add("X-Datadome-isbot", "1")
			resp.Header.Add("X-Datadome-Obiwan", "Kenoby")
			return resp, err
		},
	)

	client := &http.Client{}
	request, _ := http.NewRequest(http.MethodPost, "/validate-request", nil)
	response, _ := client.Do(request)

	addDataDomeRequestHeaders(response, request)

	assert.Equal(t, "1", request.Header.Get("X-Datadome-isbot"))
	assert.Equal(t, "", request.Header.Get("X-DataDome-Obiwan"))
}
