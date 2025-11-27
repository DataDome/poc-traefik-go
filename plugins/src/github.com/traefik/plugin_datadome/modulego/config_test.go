package modulego

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests

func TestWithEndpoint(t *testing.T) {
	endpoint := "api.example.org"
	client, err := NewClient(
		"your-api-key",
		WithEndpoint(endpoint),
	)

	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.Equal(t, endpoint, client.Endpoint)
}

func TestWithGraphQLSupport(t *testing.T) {
	enableGraphQLSupport := true
	client, err := NewClient(
		"your-api-key",
		WithGraphQLSupport(enableGraphQLSupport),
	)

	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.Equal(t, enableGraphQLSupport, client.EnableGraphQLSupport)
}

func TestWithLogger(t *testing.T) {
	mockLogger := &MockLogger{}

	client, err := NewClient(
		"your-api-key",
		WithLogger(mockLogger),
	)

	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.Equal(t, mockLogger, client.Logger)

	client.Logger.Debug("Testing Debug")
	assert.Equal(t, "DEBUG", mockLogger.lastMessage)
	client.Logger.Info("Testing Info")
	assert.Equal(t, "INFO", mockLogger.lastMessage)
	client.Logger.Warn("Testing Warn")
	assert.Equal(t, "WARN", mockLogger.lastMessage)
	client.Logger.Error("Testing Error")
	assert.Equal(t, "ERROR", mockLogger.lastMessage)
}

func TestWithMaximumBodySize(t *testing.T) {
	t.Run("With a positive integer", func(t *testing.T) {
		maximumBodySize := 10 * 10
		client, err := NewClient(
			"your-api-key",
			WithMaximumBodySize(maximumBodySize),
		)

		assert.NotNil(t, client)
		assert.Nil(t, err)
		assert.Equal(t, maximumBodySize, client.MaximumBodySize)
	})

	t.Run("With an integer less than or equal to 0", func(t *testing.T) {
		maximumBodySize := 0
		client, err := NewClient(
			"your-api-key",
			WithMaximumBodySize(maximumBodySize),
		)

		assert.Nil(t, client)
		assert.NotNil(t, err)
		assert.Equal(t, "MaximumBodySize must be a positive integer", err.Error())
	})
}

func TestWithReferrerRestoration(t *testing.T) {
	enableReferrerRestoration := true
	client, err := NewClient(
		"your-api-key",
		WithReferrerRestoration(enableReferrerRestoration),
	)

	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.Equal(t, enableReferrerRestoration, client.EnableReferrerRestoration)
}

func TestWithTimeout(t *testing.T) {
	t.Run("With a positive integer", func(t *testing.T) {
		timeout := 1500
		client, err := NewClient(
			"your-api-key",
			WithTimeout(timeout),
		)

		assert.NotNil(t, client)
		assert.Nil(t, err)
		assert.Equal(t, timeout, client.Timeout)
	})

	t.Run("With an integer less than or equal to 0", func(t *testing.T) {
		timeout := 0
		client, err := NewClient(
			"your-api-key",
			WithTimeout(timeout),
		)

		assert.Nil(t, client)
		assert.NotNil(t, err)
		assert.Equal(t, "Timeout must be a positive integer", err.Error())
	})
}

func TestWithUrlPatternExclusion(t *testing.T) {
	t.Run("With a valid RegExp", func(t *testing.T) {
		urlPatternExclusion := `(?i)\/excluded-path\/.*`
		client, err := NewClient(
			"your-api-key",
			WithUrlPatternExclusion(urlPatternExclusion),
		)

		assert.NotNil(t, client)
		assert.Nil(t, err)
		assert.Equal(t, urlPatternExclusion, client.UrlPatternExclusion)
		assert.NotNil(t, client.urlPatternExclusion)
	})

	t.Run("With an invalid RegExp", func(t *testing.T) {
		urlPatternExclusion := `(?i)\/excluded-path\/with-[error.*`
		client, err := NewClient(
			"your-api-key",
			WithUrlPatternExclusion(urlPatternExclusion),
		)

		assert.Nil(t, client)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "UrlPatternExclusion must be a valid RegExp"))
	})
}

func TestWithUrlPatternInclusion(t *testing.T) {
	t.Run("With a valid RegExp", func(t *testing.T) {
		urlPatternInclusion := `(?i)\/included-path\/.*`
		client, err := NewClient(
			"your-api-key",
			WithUrlPatternInclusion(urlPatternInclusion),
		)

		assert.NotNil(t, client)
		assert.Nil(t, err)
		assert.Equal(t, urlPatternInclusion, client.UrlPatternInclusion)
		assert.NotNil(t, client.urlPatternInclusion)
	})

	t.Run("With an invalid RegExp", func(t *testing.T) {
		urlPatternInclusion := `(?i)\/included-path\/with-[error.*`
		client, err := NewClient(
			"your-api-key",
			WithUrlPatternInclusion(urlPatternInclusion),
		)

		assert.Nil(t, client)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "UrlPatternInclusion must be a valid RegExp"))
	})
}

func TestWithXForwardedHost(t *testing.T) {
	client, err := NewClient(
		"your-api-key",
		WithXForwardedHost(true),
	)

	assert.NotNil(t, client)
	assert.Nil(t, err)
	assert.True(t, client.UseXForwardedHost)
}

// Testable examples

func ExampleWithEndpoint() {
	c, _ := NewClient("your-api-key", WithEndpoint("api.example.org"))

	fmt.Println(c.Endpoint)
	// Output: api.example.org
}

func ExampleWithGraphQLSupport() {
	c, _ := NewClient("your-api-key", WithGraphQLSupport(true))

	fmt.Println(c.EnableGraphQLSupport)
	// Output: true
}

func ExampleWithLogger() {
	mockLogger := &MockLogger{}
	c, _ := NewClient("your-api-key", WithLogger(mockLogger))

	c.Logger.Info("test")
	// Output: INFO test
}

func ExampleWithMaximumBodySize() {
	c, _ := NewClient("your-api-key", WithMaximumBodySize(10*10))

	fmt.Println(c.MaximumBodySize)
	// Output: 100
}

func ExampleWithReferrerRestoration() {
	c, _ := NewClient("your-api-key", WithReferrerRestoration(true))

	fmt.Println(c.EnableReferrerRestoration)
	// Output: true
}

func ExampleWithTimeout() {
	c, _ := NewClient("your-api-key", WithTimeout(300))

	fmt.Println(c.Timeout)
	// Output: 300
}

func ExampleWithUrlPatternExclusion() {
	c, _ := NewClient("your-api-key", WithUrlPatternExclusion(`(?i)\/not-this-path\/.*`))

	fmt.Println(c.UrlPatternExclusion)
	// Output: (?i)\/not-this-path\/.*
}

func ExampleWithUrlPatternInclusion() {
	c, _ := NewClient("your-api-key", WithUrlPatternInclusion(`(?i)\/this-path\/.*`))

	fmt.Println(c.urlPatternInclusion)
	// Output: (?i)\/this-path\/.*
}

func ExampleWithXForwardedHost() {
	c, _ := NewClient("your-api-key", WithXForwardedHost(true))
	fmt.Println(c.UseXForwardedHost)
	// Output: true
}
