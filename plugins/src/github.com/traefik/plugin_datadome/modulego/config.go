package modulego

type Option func(*Client)

// WithEndpoint is a functional option to set the endpoint of the Protection API.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.Endpoint = endpoint
	}
}

// WithGraphQLSupport is a functional option to enable the GraphQL support.
func WithGraphQLSupport(enableGraphQLSupport bool) Option {
	return func(c *Client) {
		c.EnableGraphQLSupport = enableGraphQLSupport
	}
}

// WithLogger is a functional option to set a custom Logger for the Client.
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithMaximumBodySize is a functional option to set the maximum size of a body to be analyzed.
// This option may be set when the GraphQL Support is enabled.
func WithMaximumBodySize(maximumBodySize int) Option {
	return func(c *Client) {
		c.MaximumBodySize = maximumBodySize
	}
}

// WithReferrerRestoration is a functional option to enable the referrer restoration feature.
func WithReferrerRestoration(enableReferrerRestoration bool) Option {
	return func(c *Client) {
		c.EnableReferrerRestoration = enableReferrerRestoration
	}
}

// WithTimeout is a functional option to set the HTTP Client timeout in milliseconds.
func WithTimeout(timeout int) Option {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithUrlPatternExclusion is a functional option to define the regular expression to exclude the request from being processed with the Protection API.
func WithUrlPatternExclusion(urlPatternExclusion string) Option {
	return func(c *Client) {
		c.UrlPatternExclusion = urlPatternExclusion
	}
}

// WithUrlPatternInclusion is a functional option to define the regular expression to match to process the request with the Protection API.
func WithUrlPatternInclusion(urlPatternInclusion string) Option {
	return func(c *Client) {
		c.UrlPatternInclusion = urlPatternInclusion
	}
}

// WithXForwardedHost is a functional option to indicate to use the X-Forwarded-Host header first.
func WithXForwardedHost(useXForwardedHost bool) Option {
	return func(c *Client) {
		c.UseXForwardedHost = useXForwardedHost
	}
}
