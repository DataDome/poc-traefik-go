package modulego

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultEnableGraphQLSupportValue      = false
	DefaultEnableReferrerRestorationValue = false
	DefaultEndpointValue                  = "api.datadome.co"
	DefaultMaximumBodySizeValue           = 25 * 1024
	DefaultModuleNameValue                = "Golang"
	DefaultModuleVersionValue             = "2.2.0"
	DefaultTimeoutValue                   = 150
	DefaultUrlPatternInclusionValue       = ""
	DefaultUrlPatternExclusionValue       = `(?i)\.(avi|avif|bmp|css|eot|flac|flv|gif|gz|ico|jpeg|jpg|js|json|less|map|mka|mkv|mov|mp3|mp4|mpeg|mpg|ogg|ogm|opus|otf|png|svg|svgz|swf|ttf|wav|webm|webp|woff|woff2|xml|zip)$`
	DefaultUseXForwardedHostValue         = false
)

// Client is used to interract with the DataDome's Protection API.
// This structure contains all the informations specified through the [Option]'s functions.
type Client struct {
	EnableGraphQLSupport      bool
	EnableReferrerRestoration bool
	Endpoint                  string
	Logger                    Logger
	MaximumBodySize           int
	ModuleName                string
	ModuleVersion             string
	ServerSideKey             string
	Timeout                   int
	UrlPatternInclusion       string
	UrlPatternExclusion       string
	UseXForwardedHost         bool

	endpoint            string
	httpClient          *http.Client
	urlPatternExclusion *regexp.Regexp
	urlPatternInclusion *regexp.Regexp
}

// NewClient instantiate a new DataDome [Client] to perform calls to Protection API.
// The fields may be customized through [Option] functions.
// It returns an error in case of [incorrect / invalid] inputs in the options.
func NewClient(serverSideKey string, options ...Option) (*Client, error) {
	c := &Client{
		EnableGraphQLSupport:      DefaultEnableGraphQLSupportValue,
		EnableReferrerRestoration: DefaultEnableReferrerRestorationValue,
		Endpoint:                  DefaultEndpointValue,
		Logger:                    NewDefaultLogger(),
		MaximumBodySize:           DefaultMaximumBodySizeValue,
		ModuleName:                DefaultModuleNameValue,
		ModuleVersion:             DefaultModuleVersionValue,
		ServerSideKey:             serverSideKey,
		Timeout:                   DefaultTimeoutValue,
		UrlPatternInclusion:       DefaultUrlPatternInclusionValue,
		UrlPatternExclusion:       DefaultUrlPatternExclusionValue,
	}

	// apply functional options
	for _, opt := range options {
		opt(c)
	}

	// error management
	if c.ServerSideKey == "" {
		return nil, fmt.Errorf("ServerSideKey must be defined")
	}
	if c.Timeout <= 0 {
		return nil, fmt.Errorf("Timeout must be a positive integer")
	}
	if c.MaximumBodySize <= 0 {
		return nil, fmt.Errorf("MaximumBodySize must be a positive integer")
	}

	// set not exported values
	c.httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Millisecond * time.Duration(c.Timeout),
	}
	if c.UrlPatternExclusion != "" {
		r, err := regexp.Compile(c.UrlPatternExclusion)
		if err != nil {
			return nil, fmt.Errorf("UrlPatternExclusion must be a valid RegExp: %w", err)
		}
		c.urlPatternExclusion = r
	}
	if c.UrlPatternInclusion != "" {
		r, err := regexp.Compile(c.UrlPatternInclusion)
		if err != nil {
			return nil, fmt.Errorf("UrlPatternInclusion must be a valid RegExp: %w", err)
		}
		c.urlPatternInclusion = r
	}
	c.endpoint = c.Endpoint
	if !strings.HasPrefix(c.Endpoint, "http") && !strings.HasPrefix(c.Endpoint, "/") {
		c.endpoint = fmt.Sprintf("https://%s/validate-request", c.Endpoint)
	}

	return c, nil
}

// handler is used to validate incoming requests
// This function will:
// 1. Verifies the request URL does not match the UrlPatternExclusion
// 2. Verifies the request URL match the UrlPatternInclusion (if set)
// 3. Builds the request payload for the Protection API
// 4. Performs the call to the Protection API and interpret the response
func (c *Client) handler(w http.ResponseWriter, r *http.Request, next http.Handler) (bool, error) {
	sendNext := func(res bool, err error, response http.ResponseWriter) (bool, error) {
		if next != nil {
			next.ServeHTTP(response, r)
		} else {
			return res, err
		}
		return res, nil
	}

	uri := getURI(r)
	// Test exclusion regex
	if c.urlPatternExclusion != nil && c.urlPatternExclusion.MatchString(uri) {
		c.Logger.Info("UrlPatternExclusion matches requested URI, skipping.")
		return false, nil
	}

	// Test inclusion regex
	if c.urlPatternInclusion != nil && !c.urlPatternInclusion.MatchString(uri) {
		c.Logger.Info("UrlPatternInclusion does not match requested URI, skipping.")
		return false, nil
	}

	queryStr, err := c.buildRequest(r)
	if err != nil {
		c.Logger.Error("error when building request payload: %v", err)
		return sendNext(false, err, w)
	}

	err, resp, isBlocked := c.datadomeCall(queryStr, r, w)
	if err != nil {
		c.Logger.Error("error when performing call to Protection API: %v", err)
		return sendNext(isBlocked, err, resp)
	}
	return sendNext(isBlocked, nil, resp)
}

// DatadomeHandler implements the [http.Handler] interface
func (c *Client) DatadomeHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := c.handler(w, r, next)
		if err != nil {
			panic(err)
		}
	})
}

// DatadomeProtect validates the incoming request
func (c *Client) DatadomeProtect(rw http.ResponseWriter, r *http.Request) (isBlocked bool, err error) {
	return c.handler(rw, r, nil)
}

// buildRequest extracts information from the request and build the payload to be sent to the Protection API.
// An error may be returned if the IP cannot be retrieved or if it fails to URL-encode the payload.
func (c *Client) buildRequest(r *http.Request) (string, error) {
	// Build DataDome request with the original request
	contentLength := "0"
	if r.Header.Get("content-length") != "" {
		contentLength = r.Header.Get("content-length")
	}

	authorizationLen := "0"
	if r.Header.Get("authorization") != "" {
		authorizationLen = strconv.Itoa(len(r.Header.Get("authorization")))
	}

	proto := getProtocol(r)

	port := r.URL.Port()
	if port == "" {
		if proto == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	host := r.Host
	if c.UseXForwardedHost {
		host = getHost(r)
	}

	cookiesLen := "0"
	if r.Header.Get("Cookie") != "" {
		cookiesLen = strconv.Itoa(len(r.Header.Get("cookie")))
	}

	cookiesList := getCookieList(r)

	ip, err := getIP(r)
	if err != nil {
		return "", fmt.Errorf("fail to parse request's IP: %w", err)
	}

	if c.EnableReferrerRestoration {
		isMatching, err := isMatchingReferrer(r)
		if err != nil {
			c.Logger.Warn("fail to check if the referrer matches: %v", err)
		} else if isMatching {
			err = restoreReferrer(r)
			if err != nil {
				c.Logger.Warn("fail to restore the referrer: %v", err)
			}
		}
	}

	ddRequestParams := ProtectionAPIRequestPayload{
		Key:                    c.ServerSideKey,
		IP:                     ip,
		Accept:                 truncateValue(Accept, r.Header.Get("accept")),
		AcceptCharset:          truncateValue(AcceptCharset, r.Header.Get("accept-charset")),
		AcceptEncoding:         truncateValue(AcceptEncoding, r.Header.Get("accept-encoding")),
		AcceptLanguage:         truncateValue(AcceptLanguage, r.Header.Get("accept-language")),
		APIConnectionState:     "new",
		AuthorizationLen:       authorizationLen,
		CacheControl:           truncateValue(CacheControl, r.Header.Get("cache-control")),
		ClientID:               truncateValue(ClientID, getClientId(r)),
		Connection:             truncateValue(Connection, r.Header.Get("connection")),
		ContentType:            truncateValue(ContentType, r.Header.Get("content-type")),
		CookiesLen:             cookiesLen,
		CookiesList:            cookiesList,
		From:                   truncateValue(From, r.Header.Get("from")),
		HeadersList:            truncateValue(HeadersList, getHeaderList(r)),
		Host:                   truncateValue(Host, host),
		Method:                 r.Method,
		ModuleVersion:          c.ModuleVersion,
		Origin:                 truncateValue(Origin, r.Header.Get("origin")),
		Port:                   port,
		PostParamLen:           contentLength,
		Pragma:                 truncateValue(Pragma, r.Header.Get("pragma")),
		Protocol:               proto,
		Referer:                truncateValue(Referer, r.Header.Get("referer")),
		Request:                truncateValue(Request, getURL(r)),
		RequestModuleName:      c.ModuleName,
		SecChDeviceMemory:      truncateValue(SecCHDeviceMemory, r.Header.Get("sec-ch-device-memory")),
		SecChUA:                truncateValue(SecCHUA, r.Header.Get("sec-ch-ua")),
		SecChUAArch:            truncateValue(SecCHUAArch, r.Header.Get("sec-ch-ua-arch")),
		SecChUAFullVersionList: truncateValue(SecCHUAFullVersionList, r.Header.Get("sec-ch-ua-full-version-list")),
		SecChUAMobile:          truncateValue(SecCHUAMobile, r.Header.Get("sec-ch-ua-mobile")),
		SecChUAModel:           truncateValue(SecCHUAModel, r.Header.Get("sec-ch-ua-model")),
		SecChUAPlatform:        truncateValue(SecCHUAPlatform, r.Header.Get("sec-ch-ua-platform")),
		SecFetchDest:           truncateValue(SecFetchDest, r.Header.Get("sec-fetch-dest")),
		SecFetchMode:           truncateValue(SecFetchMode, r.Header.Get("sec-fetch-mode")),
		SecFetchSite:           truncateValue(SecFetchSite, r.Header.Get("sec-fetch-site")),
		SecFetchUser:           truncateValue(SecFetchUser, r.Header.Get("sec-fetch-user")),
		ServerHostName:         truncateValue(ServerHostname, host),
		ServerName:             truncateValue(ServerName, host),
		TimeRequest:            getMicroTime(),
		TrueClientIP:           truncateValue(TrueClientIP, r.Header.Get("true-client-ip")),
		UserAgent:              truncateValue(UserAgent, r.Header.Get("user-agent")),
		Via:                    truncateValue(Via, r.Header.Get("via")),
		XForwardedForIP:        truncateValue(XForwardedForIP, r.Header.Get("x-forwarded-for")),
		XRealIP:                truncateValue(XRealIP, r.Header.Get("x-real-ip")),
		XRequestedWith:         truncateValue(XRequestedWith, r.Header.Get("x-requested-with")),
	}

	if c.EnableGraphQLSupport && isGraphQLRequest(r) {
		gqlData, err := getGraphQLData(r, c.MaximumBodySize)
		if err != nil {
			c.Logger.Warn("fail to retrieve GraphQL data: %v", err)
		}
		if gqlData != nil && gqlData.Count != 0 {
			operationName := truncateValue(GraphQLOperationName, gqlData.Name)
			ddRequestParams.GraphQLOperationName = &operationName
			ddRequestParams.GraphQLOperationType = gqlData.Type
			ddRequestParams.GraphQLOperationCount = strconv.Itoa(gqlData.Count)
		}
	}

	return ddRequestParams.ToURLValues().Encode(), nil
}

// datadomeCall performs a request to the Protection API
func (c *Client) datadomeCall(jsonStr string, origReq *http.Request, origResp http.ResponseWriter) (err error, rw http.ResponseWriter, isBlocked bool) {
	body := strings.NewReader(jsonStr)
	req, err := http.NewRequestWithContext(origReq.Context(), "POST", c.endpoint, body)
	if err != nil {
		return fmt.Errorf("error when instancing new DataDome request %w", err), nil, false
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("user-agent", "DataDome")

	if origReq.Header.Get("x-datadome-clientid") != "" {
		req.Header.Set("x-datadome-x-set-cookie", "true")
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error when performing DataDome request: %w", err), nil, false
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error when reading DataDome response %w", err), nil, false
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.Logger.Warn("error when closing the Body: %v", err)
		}
	}(response.Body)

	ddStatus := response.Header.Get("x-datadomeresponse")
	ddRespStatus := strconv.Itoa(response.StatusCode)

	if ddStatus == "" || (ddRespStatus != ddStatus) {
		c.Logger.Debug("fail to get status code and response headers from Protection API response. reason: %s", string(responseBody))
		return fmt.Errorf("fails to get status code and response headers from Protection API response. Bypass DataDome. Full DataDome response: %v", response), nil, false
	}

	// Handler DataDome status code
	switch ddStatus {
	case "400":
		return nil, origResp, false
	case "301", "302", "401", "403":
		origResp = addDataDomeHeaders(response, origResp)
		origResp.WriteHeader(response.StatusCode)
		_, err = origResp.Write(responseBody)
		if err != nil {
			return err, nil, false
		}
		return nil, origResp, true

	case "200":
		addDataDomeRequestHeaders(response, origReq)
		origResp = addDataDomeHeaders(response, origResp)
		return nil, origResp, false

	default:
		return fmt.Errorf("%s response from Protection API - Unexpected error. If the error remains, please contact us at support@datadome.co. Full response: %v", ddStatus, response.Header), origResp, false
	}
}

// addDataDomeRequestHeaders add the headers listed in the `X-datadome-request-headers`
// header of the Protection API response to the original request.
func addDataDomeRequestHeaders(ddResp *http.Response, origReq *http.Request) {
	datadomeHeadersStr := ddResp.Header.Get("x-datadome-request-headers")
	if datadomeHeadersStr != "" {
		datadomeHeaders := strings.Fields(datadomeHeadersStr)
		for _, datadomeHeaderName := range datadomeHeaders {
			datadomeHeaderValue := ddResp.Header.Get(datadomeHeaderName)
			if datadomeHeaderValue != "" {
				origReq.Header.Add(datadomeHeaderName, datadomeHeaderValue)
			}
		}
	}
}

// addDataDomeHeaders add the headers listed in the `x-datadome-headers` header
// of the Protection API response to the original response.
func addDataDomeHeaders(ddResp *http.Response, origResp http.ResponseWriter) http.ResponseWriter {
	datadomeHeadersStr := ddResp.Header.Get("x-datadome-headers")
	if datadomeHeadersStr != "" {
		datadomeHeaders := strings.Fields(datadomeHeadersStr)
		for _, datadomeHeaderName := range datadomeHeaders {
			datadomeHeaderValue := ddResp.Header.Get(datadomeHeaderName)
			if datadomeHeaderValue != "" {
				if strings.EqualFold(datadomeHeaderName, "set-cookie") {
					origResp.Header().Add(datadomeHeaderName, datadomeHeaderValue)
				} else {
					origResp.Header().Set(datadomeHeaderName, datadomeHeaderValue)
				}
			}
		}
	}
	return origResp
}
