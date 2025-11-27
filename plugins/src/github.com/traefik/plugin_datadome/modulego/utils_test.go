package modulego

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.RemoteAddr = "127.0.0.1:1234"
	request.Header.Set("Hello", "World")
	request.Header.Set("X-Test", "123")
	request.AddCookie(&http.Cookie{
		Name:  "Foo",
		Value: "fake_value1",
	})
	request.AddCookie(&http.Cookie{
		Name:  "Bar",
		Value: "fake_value2",
	})

	return request
}

func TestMicroTime(t *testing.T) {
	if len(getMicroTime()) != 16 {
		t.Error("Microtime unit test fail")
		return
	}
}

func TestGetIP(t *testing.T) {
	request := setup()

	result, err := getIP(request)
	assert.Equal(t, "127.0.0.1", result)
	assert.Equal(t, nil, err)
}

func TestGetClientId(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/this-is-the-way", nil)
	req.Header.Set("x-datadome-clientid", "123456")

	result := getClientId(req)

	assert.Equal(t, "123456", result)
}

func TestGetCookieList(t *testing.T) {
	request := setup()

	result := getCookieList(request)

	assert.Contains(t, result, "Foo")
	assert.Contains(t, result, "Bar")
}

func TestGetHeaderList(t *testing.T) {
	request := setup()

	result := getHeaderList(request)
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "X-Test")
}

func TestGetURL(t *testing.T) {
	request := setup()

	result := getURL(request)
	assert.Equal(t, "/ping", result)

	request = httptest.NewRequest(http.MethodGet, "/ping?a=b", nil)
	result = getURL(request)
	assert.Equal(t, "/ping?a=b", result)
}

func TestGetURI(t *testing.T) {
	tests := []struct {
		want  string
		input *http.Request
	}{
		{want: "example.com/ping", input: httptest.NewRequest(http.MethodGet, "http://example.com/ping", nil)},
		{want: "example.com", input: httptest.NewRequest(http.MethodGet, "http://example.com", nil)},
		{want: "example.com/ping", input: httptest.NewRequest(http.MethodGet, "http://example.com/ping?foo=bar", nil)},
		{want: "example.com/ping", input: httptest.NewRequest(http.MethodGet, "http://example.com/ping#fragment", nil)},
		{want: "example.com/ping", input: httptest.NewRequest(http.MethodGet, "http://example.com/ping%3Fencoded%3DqueryParams", nil)},
		{want: "example.com/ping", input: httptest.NewRequest(http.MethodGet, "http://example.com/ping%23encodedFragment", nil)},
	}

	for _, tc := range tests {
		got := getURI(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

// TestParseGraphQLQuery with different type of GraphQL JSON body
// 1. Shorthand Syntax
// 2. Mutation
// 3. Query
// 4. Multiple operations
// 5. Wrong GraphQL query
func TestParseGraphQLQuery(t *testing.T) {
	tests := []struct {
		want  GraphQLData
		input string
	}{
		{want: GraphQLData{Count: 1, Name: "", Type: Query}, input: `{"query":"{ todos { title }}"}`},
		{want: GraphQLData{Count: 1, Name: "LoginV2", Type: Mutation}, input: `{"query":"mutation LoginV2($loginV2LoginName2: String!, $loginV2Password2: String!, $loginV2ServiceLocationId3: ID!) { loginV2(loginName: $loginV2LoginName2, password: $loginV2Password2, serviceLocationId: $loginV2ServiceLocationId3) { cookieAdapterToken msg }}","variables":{"loginV2LoginName2":"alpha@staging.com","loginV2Password2":"Test1234","loginV2ServiceLocationId3":"50000059"}}`},
		{want: GraphQLData{Count: 1, Name: "Coupons", Type: Query}, input: `{"query":"query Coupons($couponsServiceLocationId3: ID!) { coupons(serviceLocationId: $couponsServiceLocationId3) { coupons { name title circularId couponType endDate } } }","variables":{"couponsServiceLocationId3":"50000059"}}`},
		{want: GraphQLData{Count: 4, Name: "One", Type: Mutation}, input: `{"query":"mutation One { # Do something ...}mutation Two { # Do something ...}query Three @depends(on: [\"One\", \"Two\"]) { # Do something ...}query Four @depends(on: \"Three\") { # Do something ...}"}`},
		{want: GraphQLData{Count: 1, Name: "", Type: Query}, input: `{"query":"query { todos { title }}"}`},
		{want: GraphQLData{Count: 0, Name: "", Type: Query}, input: `{"query":"mutatio RefreshTokenV2 $serviceLocationId: ID!) { refreshTokenV2(serviceLocationId: $serviceLocationId) { cookieAdapterToken msg }}","variables":{"serviceLocationId":""}}`},
	}

	for _, tc := range tests {
		got := parseGraphQLQuery(tc.input)
		assert.Equal(t, tc.want.Count, got.Count)
		assert.Equal(t, tc.want.Name, got.Name)
		assert.Equal(t, tc.want.Type, got.Type)
	}
}

func TestTruncateValue(t *testing.T) {
	type Header struct {
		Key   ApiFields
		Value string
	}
	fakeCommonValue := strings.Repeat("a", 3000)
	fakeEndXFFValue := strings.Repeat("b", 512)
	fakeXFFValue := fakeCommonValue + fakeEndXFFValue

	tests := []struct {
		want  int
		input Header
	}{
		{want: 8, input: Header{Key: SecFetchUser, Value: fakeCommonValue}},
		{want: 16, input: Header{Key: SecCHUAArch, Value: fakeCommonValue}},
		{want: 32, input: Header{Key: SecFetchDest, Value: fakeCommonValue}},
		{want: 64, input: Header{Key: ContentType, Value: fakeCommonValue}},
		{want: 128, input: Header{Key: SecCHUA, Value: fakeCommonValue}},
		{want: 256, input: Header{Key: AcceptLanguage, Value: fakeCommonValue}},
		{want: 512, input: Header{Key: Origin, Value: fakeCommonValue}},
		{want: 768, input: Header{Key: UserAgent, Value: fakeCommonValue}},
		{want: 1024, input: Header{Key: Referer, Value: fakeCommonValue}},
		{want: 2048, input: Header{Key: Request, Value: fakeCommonValue}},
		{want: 3000, input: Header{Key: "RequestModuleName", Value: fakeCommonValue}},
		{want: 512, input: Header{Key: XForwardedForIP, Value: fakeXFFValue}},
		{want: 0, input: Header{Key: "SomeHeader", Value: ""}},
	}

	for _, tc := range tests {
		got := truncateValue(tc.input.Key, tc.input.Value)
		assert.Equal(t, tc.want, len(got))
		if tc.input.Key == XForwardedForIP {
			assert.Equal(t, fakeEndXFFValue, got)
		}
	}
}

func TestGetProtocol(t *testing.T) {
	reqHTTP := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	reqHTTPSWithTLS := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	reqHTTPWithXFP := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	reqHTTPSWithXFP := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	reqHTTPSWithTLS.TLS = &tls.ConnectionState{}
	reqHTTPWithXFP.Header.Set("X-Forwarded-Proto", "http")
	reqHTTPSWithXFP.Header.Set("X-Forwarded-Proto", "https")

	tests := []struct {
		want  string
		input *http.Request
	}{
		{want: "http", input: reqHTTP},
		{want: "http", input: reqHTTPWithXFP},
		{want: "https", input: reqHTTPSWithTLS},
		{want: "https", input: reqHTTPSWithXFP},
	}

	for _, tc := range tests {
		got := getProtocol(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

func TestIsMatchingReferer(t *testing.T) {
	reqWithoutReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo?dd_referrer=", nil)
	reqWithNotMatchingReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo?dd_referrer=", nil)
	reqWithDDReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo?dd_referrer=http%3A%2F%2Fexample.com%2Ffoo", nil)
	reqWithMultipleQueryParams := httptest.NewRequest(http.MethodGet, "http://example.com/foo?toto=tata&dd_referrer=&foo=bar", nil)

	reqWithNotMatchingReferrer.Header.Set("Referer", "http%3A%2F%2Fhttpbin.org%2Fbar")
	reqWithDDReferrer.Header.Set("Referer", "http%3A%2F%2Fexample.com%2Ffoo")
	reqWithMultipleQueryParams.Header.Set("Referer", "http%3A%2F%2Fexample.com%2Ffoo%3Ftoto%3Dtata%26foo%3Dbar")

	tests := []struct {
		want  bool
		input *http.Request
	}{
		{want: false, input: reqWithoutReferrer},
		{want: false, input: reqWithNotMatchingReferrer},
		{want: true, input: reqWithDDReferrer},
		{want: true, input: reqWithMultipleQueryParams},
	}

	for _, tc := range tests {
		got, _ := isMatchingReferrer(tc.input)
		assert.Equal(t, tc.want, got)
	}
}

func TestRestoreReferrer(t *testing.T) {
	type ExpectedResult struct {
		Error    error
		Referrer string
	}

	reqWithoutDDReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	reqWithFilledDDReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo?dd_referrer=Foo", nil)
	reqWithEmptyDDReferrer := httptest.NewRequest(http.MethodGet, "http://example.com/foo?dd_referrer=", nil)
	reqWithMultipleQueryParams := httptest.NewRequest(http.MethodGet, "http://example.com/foo?query=params&dd_referrer=Foo&toto=tata", nil)

	reqWithoutDDReferrer.Header.Set("Referer", "Bar")
	reqWithFilledDDReferrer.Header.Set("Referer", "Bar")
	reqWithEmptyDDReferrer.Header.Set("Referer", "Bar")
	reqWithMultipleQueryParams.Header.Set("Referer", "Bar")

	tests := []struct {
		want  ExpectedResult
		input *http.Request
	}{
		{want: ExpectedResult{Error: nil, Referrer: "Bar"}, input: reqWithoutDDReferrer},
		{want: ExpectedResult{Error: nil, Referrer: "Foo"}, input: reqWithFilledDDReferrer},
		{want: ExpectedResult{Error: nil, Referrer: ""}, input: reqWithEmptyDDReferrer},
		{want: ExpectedResult{Error: nil, Referrer: "Foo"}, input: reqWithMultipleQueryParams},
	}

	for _, tc := range tests {
		got := restoreReferrer(tc.input)
		assert.Equal(t, tc.want.Error, got)
		assert.Equal(t, tc.input.Header.Get("Referer"), tc.want.Referrer)

		tcQuery := tc.input.URL.Query()
		assert.Equal(t, "", tcQuery.Get("dd_referrer"))
	}

	// Test case for multiple query params
	multipleQueryParams := reqWithMultipleQueryParams.URL.Query()
	assert.Equal(t, "params", multipleQueryParams.Get("query"))
	assert.Equal(t, "tata", multipleQueryParams.Get("toto"))
}

func TestGetHost(t *testing.T) {
	reqWithXForwardedHost := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	reqWithXForwardedHost.Header.Set("X-Forwarded-Host", "proxy.example.com")

	reqWithoutXForwardedHost := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	tests := []struct {
		name     string
		request  *http.Request
		expected string
	}{
		{
			name:     "With X-Forwarded-Host header",
			request:  reqWithXForwardedHost,
			expected: "proxy.example.com",
		},
		{
			name:     "Without X-Forwarded-Host header",
			request:  reqWithoutXForwardedHost,
			expected: "example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getHost(tc.request)
			assert.Equal(t, tc.expected, got)
		})
	}
}
