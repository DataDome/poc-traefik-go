package modulego

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// getMicroTime returns the current unix timestamp in microseconds
// This function needs to implement the time.UnixMicro function when Tyk will support the go version >= 1.18
func getMicroTime() string {
	return strconv.FormatInt(time.Now().UnixNano()/1000, 10)
}

// getIP returns the IP of the emitter from the RemoteAddr field of the request.
func getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}

// getClientId retrieves the ClientID from the incoming request.
// It uses the value of the `X-DataDome-ClientID` if the session by header feature is used.
// It reads the `DataDome` cookie value otherwise.
func getClientId(r *http.Request) string {
	clientIDHeaders := r.Header.Get("x-datadome-clientid")
	if len(clientIDHeaders) > 0 {
		return clientIDHeaders
	}

	cookie, err := r.Cookie("datadome")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// getCookieList returns the list of cookies' keys separated by commas
func getCookieList(r *http.Request) string {
	cookies := r.Cookies()
	cookiesList := make([]string, 0, len(cookies))

	for _, cookie := range cookies {
		cookiesList = append(cookiesList, cookie.Name)
	}

	return strings.Join(cookiesList, ",")
}

// getHeaderList returns the list of header's keys separated by commas
func getHeaderList(r *http.Request) string {
	headerNames := make([]string, 0, len(r.Header))

	for headerName := range r.Header {
		headerNames = append(headerNames, headerName)
	}

	return strings.Join(headerNames, ",")
}

// getURL returns the path and the query parameters (if present) of the request
func getURL(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return r.URL.Path + "?" + r.URL.RawQuery
	} else {
		return r.URL.Path
	}
}

// cut implementation from the official package (see https://cs.opensource.google/go/go/+/master:src/strings/strings.go;l=1278)
func cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

// getURI returns the URI without the query parameters nor the Fragments.
// This function needs to implement the strings.Cut function when Tyk will support the go version >= 1.18
func getURI(r *http.Request) string {
	var finalPath string

	pathWithoutQueryParams, _, _ := cut(r.URL.Path, "?")
	finalPath, _, _ = cut(pathWithoutQueryParams, "#")
	finalUri := r.URL.Host + finalPath

	return finalUri
}

// ApiFields describes the fields expected for the [ProtectionAPIRequestPayload]
type ApiFields string

const (
	Accept                 ApiFields = "Accept"
	AcceptCharset          ApiFields = "AcceptCharset"
	AcceptEncoding         ApiFields = "AcceptEncoding"
	AcceptLanguage         ApiFields = "AcceptLanguage"
	CacheControl           ApiFields = "CacheControl"
	ClientID               ApiFields = "ClientID"
	Connection             ApiFields = "Connection"
	ContentType            ApiFields = "ContentType"
	CookiesList            ApiFields = "CookiesList"
	From                   ApiFields = "From"
	GraphQLOperationCount  ApiFields = "GraphQLOperationCount"
	GraphQLOperationName   ApiFields = "GraphQLOperationName"
	GraphQLOperationType   ApiFields = "GraphQLOperationType"
	HeadersList            ApiFields = "HeadersList"
	Host                   ApiFields = "Host"
	Origin                 ApiFields = "Origin"
	Pragma                 ApiFields = "Pragma"
	Referer                ApiFields = "Referer"
	Request                ApiFields = "Request"
	SecCHDeviceMemory      ApiFields = "SecCHDeviceMemory"
	SecCHUA                ApiFields = "SecCHUA"
	SecCHUAArch            ApiFields = "SecCHUAArch"
	SecCHUAFullVersionList ApiFields = "SecCHUAFullVersionList"
	SecCHUAMobile          ApiFields = "SecCHUAMobile"
	SecCHUAModel           ApiFields = "SecCHUAModel"
	SecCHUAPlatform        ApiFields = "SecCHUAPlatform"
	SecFetchUser           ApiFields = "SecFetchUser"
	SecFetchDest           ApiFields = "SecFetchDest"
	SecFetchMode           ApiFields = "SecFetchMode"
	SecFetchSite           ApiFields = "SecFetchSite"
	ServerHostname         ApiFields = "ServerHostname"
	ServerName             ApiFields = "ServerName"
	TrueClientIP           ApiFields = "TrueClientIP"
	UserAgent              ApiFields = "UserAgent"
	Via                    ApiFields = "Via"
	XForwardedForIP        ApiFields = "XForwardedForIP"
	XRealIP                ApiFields = "XRealIP"
	XRequestedWith         ApiFields = "XRequestedWith"
)

// getTruncationSize returns the maximal size allowed for a given [ApiFields]
func getTruncationSize(key ApiFields) int {
	switch key {
	case SecCHDeviceMemory, SecCHUAMobile, SecFetchUser:
		return 8
	case SecCHUAArch:
		return 16
	case SecCHUAPlatform, SecFetchDest, SecFetchMode:
		return 32
	case ContentType, SecFetchSite:
		return 64
	case CacheControl, ClientID, XRequestedWith, AcceptCharset, AcceptEncoding, Pragma, Connection, From, SecCHUA, SecCHUAModel, TrueClientIP, XRealIP, GraphQLOperationName:
		return 128
	case AcceptLanguage, SecCHUAFullVersionList, Via:
		return 256
	case HeadersList, Origin, ServerHostname, ServerName, Accept, Host:
		return 512
	case XForwardedForIP:
		return -512
	case UserAgent:
		return 768
	case Referer, CookiesList:
		return 1024
	case Request:
		return 2048
	}

	return 0
}

// truncateValue returns the truncated value of the given key.
// If the value does not need to be truncated, it remains unchanged.
func truncateValue(key ApiFields, value string) string {
	if value == "" {
		return ""
	}

	limit := getTruncationSize(key)
	if limit < 0 && len(value) > (-1*limit) {
		limit *= -1
		value = value[len(value)-limit:]
	} else if limit > 0 && len(value) > limit {
		value = value[:limit]
	}

	return value
}

// isGraphQLRequest indicates if the incoming request is a GraphQL request.
func isGraphQLRequest(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/json" && r.Method == "POST" && r.ContentLength > 0 && strings.Contains(r.URL.Path, "graphql")
}

// readBodyWithoutConsuming extracts the GraphQL query from the body.
// When the body has been fully read, it is restored to the original request.
func readBodyWithoutConsuming(r *http.Request, maximumBodySize int) ([]byte, error) {
	readSize := 1024
	regexp := regexp.MustCompile(`"query"\s*:\s*("(?:query|mutation|subscription)?\s*(?:[A-Za-z_][A-Za-z0-9_]*)?\s*[{(].*)`)
	limitedReader := &io.LimitedReader{
		R: r.Body,
		N: int64(maximumBodySize),
	}

	matchBuffer := make([]byte, 0, 2*readSize)
	buffer := make([]byte, readSize)
	var beginBody bytes.Buffer
	var matchedBytes []byte
	found := false

	for {
		n, err := limitedReader.Read(buffer)
		if n > 0 {
			beginBody.Write(buffer[:n])
			if !found {
				matchBuffer = append(matchBuffer, buffer[:n]...)
				if len(matchBuffer) >= 2*readSize {
					matchBuffer = matchBuffer[len(matchBuffer)-(2*readSize):]
				}
				matches := regexp.FindSubmatch(matchBuffer)
				if len(matches) > 1 {
					matchedBytes = matches[1]
					found = true
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	restBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(beginBody.Bytes()), bytes.NewReader(restBody)))

	return matchedBytes, nil
}

// parseGraphQLQuery returns a [GraphQLData] structure with information extracted from the body.
func parseGraphQLQuery(body string) *GraphQLData {
	gqlData := &GraphQLData{
		Count: 0,
		Type:  Query,
		Name:  "",
	}
	regex := regexp.MustCompile(`(?m)(?P<operationType>query|mutation|subscription)\s*(?P<operationName>[A-Za-z_][A-Za-z0-9_]*)?\s*[({@]`)
	matches := regex.FindAllStringSubmatch(body, -1)
	matchLength := len(matches)

	if matchLength > 0 {
		gqlData.Count = matchLength
		firstMatch := matches[0]
		operationTypeIndex := regex.SubexpIndex("operationType")
		operationNameIndex := regex.SubexpIndex("operationName")

		if operationTypeIndex != -1 && operationTypeIndex < len(firstMatch) {
			gqlData.Type = OperationType(firstMatch[operationTypeIndex])
		}
		if operationNameIndex != -1 && operationNameIndex < len(firstMatch) {
			gqlData.Name = firstMatch[operationNameIndex]
		}
	} else {
		shorthandSyntaxRegex := regexp.MustCompile(`"(?:query|mutation|subscription)?\s*(?:[A-Za-z_][A-Za-z0-9_]*)?\s*[{]`)
		shorthandMatches := shorthandSyntaxRegex.FindStringSubmatch(body)
		gqlData.Count = len(shorthandMatches)
	}

	return gqlData
}

// getGraphQLData reads the body to extract the GraphQL query and parse it.
// An error is returned if
// - an error happened during the lecture of the body
// - the GraphQL query was not found
func getGraphQLData(r *http.Request, maximumBodySize int) (*GraphQLData, error) {
	body, err := readBodyWithoutConsuming(r, maximumBodySize)
	if err != nil {
		return nil, fmt.Errorf("error while reading request body: %w", err)
	}
	if body == nil {
		return nil, fmt.Errorf("query not found in the request body")
	}
	bodyStr := string(body)

	return parseGraphQLQuery(bodyStr), nil
}

// getHost returns the host of the request.
// It uses the `X-Forwarded-Host` header value if the value exists.
// Otherwise, it uses the `Host` field of the request.
func getHost(r *http.Request) string {
	xfh := r.Header.Get("X-Forwarded-Host")
	if xfh != "" {
		return xfh
	}
	return r.Host
}

// getProtocol returns the protocol of the request.
// It uses the `X-Forwarded-Proto` header value if the value is correct (i.e. `http` or `https`).
// It checks the TLS field of the request afterwards.
func getProtocol(r *http.Request) string {
	proto := "http"
	xForwardedProto := r.Header.Get("X-Forwarded-Proto")
	if strings.EqualFold(xForwardedProto, "http") || strings.EqualFold(xForwardedProto, "https") {
		proto = xForwardedProto
	} else if r.TLS != nil {
		proto = "https"
	}

	return proto
}

// sortQueryParams returns the sorted query parameter's list.
func sortQueryParams(u *url.URL) string {
	queryParams := u.Query()
	sortedQuery := url.Values{}
	for key, values := range queryParams {
		sort.Strings(values)
		sortedQuery[key] = values
	}
	u.RawQuery = queryParams.Encode()
	u.Fragment = ""

	return u.String()
}

// isMatchingReferrer checks that URL of the request is equal to the value of the `Referer` header.
// The `dd_referrer` query parameter is omitted from the request's URL.
func isMatchingReferrer(r *http.Request) (bool, error) {
	fullURL := fmt.Sprintf("%s://%s%s", getProtocol(r), r.Host, getURL(r))
	parsedUrl, err := url.Parse(fullURL)
	if err != nil {
		return false, fmt.Errorf("fail to parse request URL: %w", err)
	}
	queryParams := parsedUrl.Query()
	queryParams.Del("dd_referrer")
	parsedUrl.RawQuery = queryParams.Encode()

	referer := r.Header.Get("Referer")
	if referer == "" {
		return false, nil
	}
	decodedReferer, err := url.QueryUnescape(referer)
	if err != nil {
		return false, fmt.Errorf("fail to decode referer header: %w", err)
	}
	parsedReferer, err := url.Parse(decodedReferer)
	if err != nil {
		return false, fmt.Errorf("fail to parse referer header: %w", err)
	}

	return sortQueryParams(parsedUrl) == sortQueryParams(parsedReferer), nil
}

// restoreReferrer replaces the value of the `Referer` header with the value of the `dd_referrer` query parameter.
// If the `dd_referrer` value is empty, the `Referer` header is deleted from the request.
func restoreReferrer(r *http.Request) error {
	queryParams := r.URL.Query()
	if queryParams.Has("dd_referrer") {
		ddReferrer := queryParams.Get("dd_referrer")
		if ddReferrer == "" {
			r.Header.Del("Referer")
		} else {
			decodedReferrer, err := url.QueryUnescape(ddReferrer)
			if err != nil {
				return fmt.Errorf("fail to decoded dd_referrer query value: %w", err)
			}
			r.Header.Set("Referer", decodedReferrer)
		}
		queryParams.Del("dd_referrer")
		r.URL.RawQuery = queryParams.Encode()
	}
	return nil
}
