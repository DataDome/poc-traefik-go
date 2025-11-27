package modulego

import "net/url"

// OperationType describes the expected operations values for a GraphQL query.
type OperationType string

const (
	Mutation     OperationType = "mutation"
	Query        OperationType = "query"
	Subscription OperationType = "subscription"
)

// GraphQLData describes the informations extracted from the GraphQL query
type GraphQLData struct {
	Type  OperationType
	Name  string
	Count int
}

// ProtectionAPIRequestPayload is used to construct the payload that will be send to the Protection API
type ProtectionAPIRequestPayload struct {
	Key                    string
	RequestModuleName      string
	ModuleVersion          string
	ServerName             string
	APIConnectionState     string
	IP                     string
	Port                   string
	TimeRequest            string
	Protocol               string
	Method                 string
	ServerHostName         string
	Request                string
	HeadersList            string
	Host                   string
	UserAgent              string
	Referer                string
	Accept                 string
	AcceptEncoding         string
	AcceptLanguage         string
	AcceptCharset          string
	Origin                 string
	XForwardedForIP        string
	XRequestedWith         string
	Connection             string
	Pragma                 string
	CacheControl           string
	CookiesLen             string
	CookiesList            string
	AuthorizationLen       string
	PostParamLen           string
	XRealIP                string
	ClientID               string
	SecChDeviceMemory      string
	SecChUA                string
	SecChUAArch            string
	SecChUAFullVersionList string
	SecChUAMobile          string
	SecChUAModel           string
	SecChUAPlatform        string
	SecFetchDest           string
	SecFetchMode           string
	SecFetchSite           string
	SecFetchUser           string
	Via                    string
	From                   string
	ContentType            string
	TrueClientIP           string
	GraphQLOperationCount  string
	GraphQLOperationName   *string
	GraphQLOperationType   OperationType
}

func (p *ProtectionAPIRequestPayload) ToURLValues() url.Values {
	v := url.Values{}

	// Required fields
	v.Set("Key", p.Key)
	v.Set("IP", p.IP)
	v.Set("Request", p.Request)
	v.Set("RequestModuleName", p.RequestModuleName)

	// Optional fields - only set if not empty
	if p.ModuleVersion != "" {
		v.Set("ModuleVersion", p.ModuleVersion)
	}
	if p.ServerName != "" {
		v.Set("ServerName", p.ServerName)
	}
	if p.APIConnectionState != "" {
		v.Set("APIConnectionState", p.APIConnectionState)
	}
	if p.Port != "" {
		v.Set("Port", p.Port)
	}
	if p.TimeRequest != "" {
		v.Set("TimeRequest", p.TimeRequest)
	}
	if p.Protocol != "" {
		v.Set("Protocol", p.Protocol)
	}
	if p.Method != "" {
		v.Set("Method", p.Method)
	}
	if p.ServerHostName != "" {
		v.Set("ServerHostname", p.ServerHostName)
	}
	if p.HeadersList != "" {
		v.Set("HeadersList", p.HeadersList)
	}
	if p.Host != "" {
		v.Set("Host", p.Host)
	}
	if p.UserAgent != "" {
		v.Set("UserAgent", p.UserAgent)
	}
	if p.Referer != "" {
		v.Set("Referer", p.Referer)
	}
	if p.Accept != "" {
		v.Set("Accept", p.Accept)
	}
	if p.AcceptEncoding != "" {
		v.Set("AcceptEncoding", p.AcceptEncoding)
	}
	if p.AcceptLanguage != "" {
		v.Set("AcceptLanguage", p.AcceptLanguage)
	}
	if p.AcceptCharset != "" {
		v.Set("AcceptCharset", p.AcceptCharset)
	}
	if p.Origin != "" {
		v.Set("Origin", p.Origin)
	}
	if p.XForwardedForIP != "" {
		v.Set("XForwardedForIP", p.XForwardedForIP)
	}
	if p.XRequestedWith != "" {
		v.Set("X-Requested-With", p.XRequestedWith)
	}
	if p.Connection != "" {
		v.Set("Connection", p.Connection)
	}
	if p.Pragma != "" {
		v.Set("Pragma", p.Pragma)
	}
	if p.CacheControl != "" {
		v.Set("CacheControl", p.CacheControl)
	}
	if p.CookiesLen != "" {
		v.Set("CookiesLen", p.CookiesLen)
	}
	if p.CookiesList != "" {
		v.Set("CookiesList", p.CookiesList)
	}
	if p.AuthorizationLen != "" {
		v.Set("AuthorizationLen", p.AuthorizationLen)
	}
	if p.PostParamLen != "" {
		v.Set("PostParamLen", p.PostParamLen)
	}
	if p.XRealIP != "" {
		v.Set("X-Real-IP", p.XRealIP)
	}
	if p.ClientID != "" {
		v.Set("ClientID", p.ClientID)
	}
	if p.SecChDeviceMemory != "" {
		v.Set("SecCHDeviceMemory", p.SecChDeviceMemory)
	}
	if p.SecChUA != "" {
		v.Set("SecCHUA", p.SecChUA)
	}
	if p.SecChUAArch != "" {
		v.Set("SecCHUAArch", p.SecChUAArch)
	}
	if p.SecChUAFullVersionList != "" {
		v.Set("SecCHUAFullVersionList", p.SecChUAFullVersionList)
	}
	if p.SecChUAMobile != "" {
		v.Set("SecCHUAMobile", p.SecChUAMobile)
	}
	if p.SecChUAModel != "" {
		v.Set("SecCHUAModel", p.SecChUAModel)
	}
	if p.SecChUAPlatform != "" {
		v.Set("SecCHUAPlatform", p.SecChUAPlatform)
	}
	if p.SecFetchDest != "" {
		v.Set("SecFetchDest", p.SecFetchDest)
	}
	if p.SecFetchMode != "" {
		v.Set("SecFetchMode", p.SecFetchMode)
	}
	if p.SecFetchSite != "" {
		v.Set("SecFetchSite", p.SecFetchSite)
	}
	if p.SecFetchUser != "" {
		v.Set("SecFetchUser", p.SecFetchUser)
	}
	if p.Via != "" {
		v.Set("Via", p.Via)
	}
	if p.From != "" {
		v.Set("From", p.From)
	}
	if p.ContentType != "" {
		v.Set("ContentType", p.ContentType)
	}
	if p.TrueClientIP != "" {
		v.Set("TrueClientIP", p.TrueClientIP)
	}
	if p.GraphQLOperationCount != "" {
		v.Set("GraphQLOperationCount", p.GraphQLOperationCount)
	}
	if p.GraphQLOperationName != nil {
		v.Set("GraphQLOperationName", *p.GraphQLOperationName)
	}
	if p.GraphQLOperationType != "" {
		v.Set("GraphQLOperationType", string(p.GraphQLOperationType))
	}

	return v
}
