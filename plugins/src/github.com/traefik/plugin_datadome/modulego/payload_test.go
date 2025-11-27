package modulego

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtectionAPIRequestPayload_ToURLValues_RequiredParamsOnly(t *testing.T) {
	// Arrange: Create payload with only required fields
	payload := ProtectionAPIRequestPayload{
		Key:               "test-key-123",
		IP:                "192.168.1.1",
		Request:           "/api/endpoint",
		RequestModuleName: "Go",
	}

	// Act: Convert to URL values and encode
	values := payload.ToURLValues()
	encoded := values.Encode()

	// Assert: Verify required parameters are present
	assert.Equal(t, "test-key-123", values.Get("Key"), "Key should be present")
	assert.Equal(t, "192.168.1.1", values.Get("IP"), "IP should be present")
	assert.Equal(t, "/api/endpoint", values.Get("Request"), "Request should be present")
	assert.Equal(t, "Go", values.Get("RequestModuleName"), "RequestModuleName should be present")

	// Assert: Verify encoded string contains all required params
	parsedValues, err := url.ParseQuery(encoded)
	assert.NoError(t, err, "Encoded string should be valid URL query")
	assert.Equal(t, "test-key-123", parsedValues.Get("Key"))
	assert.Equal(t, "192.168.1.1", parsedValues.Get("IP"))
	assert.Equal(t, "/api/endpoint", parsedValues.Get("Request"))
	assert.Equal(t, "Go", parsedValues.Get("RequestModuleName"))

	// Assert: Verify optional parameters are NOT present
	assert.Empty(t, values.Get("ModuleVersion"), "ModuleVersion should not be present")
	assert.Empty(t, values.Get("UserAgent"), "UserAgent should not be present")
	assert.Empty(t, values.Get("Host"), "Host should not be present")
}

func TestProtectionAPIRequestPayload_ToURLValues_AdditionalParams(t *testing.T) {
	// Arrange: Create payload with required and additional optional fields
	graphqlOpName := "getUserProfile"
	payload := ProtectionAPIRequestPayload{
		// Required fields
		Key:               "test-key-456",
		IP:                "10.0.0.1",
		Request:           "/graphql",
		RequestModuleName: "Go",

		// Additional optional fields
		ModuleVersion:          "2.2.0",
		ServerName:             "api-server-01",
		APIConnectionState:     "new",
		Port:                   "443",
		TimeRequest:            "1234567890",
		Protocol:               "HTTPS/1.1",
		Method:                 "POST",
		ServerHostName:         "api.example.com",
		HeadersList:            "accept,user-agent,content-type",
		Host:                   "api.example.com",
		UserAgent:              "Mozilla/5.0",
		Referer:                "https://example.com",
		Accept:                 "application/json",
		AcceptEncoding:         "gzip, deflate, br",
		AcceptLanguage:         "en-US,en;q=0.9",
		AcceptCharset:          "utf-8",
		Origin:                 "https://example.com",
		XForwardedForIP:        "203.0.113.1",
		XRequestedWith:         "XMLHttpRequest",
		Connection:             "keep-alive",
		Pragma:                 "no-cache",
		CacheControl:           "no-cache",
		CookiesLen:             "3",
		CookiesList:            "session,user_id,preferences",
		AuthorizationLen:       "20",
		PostParamLen:           "150",
		XRealIP:                "203.0.113.2",
		ClientID:               "client-abc-123",
		SecChDeviceMemory:      "8",
		SecChUA:                `"Chromium";v="120", "Google Chrome";v="120"`,
		SecChUAArch:            "x86",
		SecChUAFullVersionList: `"Chromium";v="120.0.6099.129"`,
		SecChUAMobile:          "?0",
		SecChUAModel:           "Pixel 7",
		SecChUAPlatform:        "Windows",
		SecFetchDest:           "empty",
		SecFetchMode:           "cors",
		SecFetchSite:           "same-origin",
		SecFetchUser:           "?1",
		Via:                    "1.1 proxy.example.com",
		From:                   "webmaster@example.com",
		ContentType:            "application/json",
		TrueClientIP:           "203.0.113.3",

		// GraphQL specific fields
		GraphQLOperationCount: "1",
		GraphQLOperationName:  &graphqlOpName,
		GraphQLOperationType:  Query,
	}

	// Act: Convert to URL values and encode
	values := payload.ToURLValues()
	encoded := values.Encode()

	// Assert: Verify all required parameters are present
	assert.Equal(t, "test-key-456", values.Get("Key"))
	assert.Equal(t, "10.0.0.1", values.Get("IP"))
	assert.Equal(t, "/graphql", values.Get("Request"))
	assert.Equal(t, "Go", values.Get("RequestModuleName"))

	// Assert: Verify additional optional parameters are present
	assert.Equal(t, "2.2.0", values.Get("ModuleVersion"))
	assert.Equal(t, "api-server-01", values.Get("ServerName"))
	assert.Equal(t, "new", values.Get("APIConnectionState"))
	assert.Equal(t, "443", values.Get("Port"))
	assert.Equal(t, "1234567890", values.Get("TimeRequest"))
	assert.Equal(t, "HTTPS/1.1", values.Get("Protocol"))
	assert.Equal(t, "POST", values.Get("Method"))
	assert.Equal(t, "api.example.com", values.Get("ServerHostname"))
	assert.Equal(t, "accept,user-agent,content-type", values.Get("HeadersList"))
	assert.Equal(t, "api.example.com", values.Get("Host"))
	assert.Equal(t, "Mozilla/5.0", values.Get("UserAgent"))
	assert.Equal(t, "https://example.com", values.Get("Referer"))
	assert.Equal(t, "application/json", values.Get("Accept"))
	assert.Equal(t, "gzip, deflate, br", values.Get("AcceptEncoding"))
	assert.Equal(t, "en-US,en;q=0.9", values.Get("AcceptLanguage"))
	assert.Equal(t, "utf-8", values.Get("AcceptCharset"))
	assert.Equal(t, "https://example.com", values.Get("Origin"))
	assert.Equal(t, "203.0.113.1", values.Get("XForwardedForIP"))
	assert.Equal(t, "XMLHttpRequest", values.Get("X-Requested-With"))
	assert.Equal(t, "keep-alive", values.Get("Connection"))
	assert.Equal(t, "no-cache", values.Get("Pragma"))
	assert.Equal(t, "no-cache", values.Get("CacheControl"))
	assert.Equal(t, "3", values.Get("CookiesLen"))
	assert.Equal(t, "session,user_id,preferences", values.Get("CookiesList"))
	assert.Equal(t, "20", values.Get("AuthorizationLen"))
	assert.Equal(t, "150", values.Get("PostParamLen"))
	assert.Equal(t, "203.0.113.2", values.Get("X-Real-IP"))
	assert.Equal(t, "client-abc-123", values.Get("ClientID"))

	// Assert: Verify Sec-CH-* headers
	assert.Equal(t, "8", values.Get("SecCHDeviceMemory"))
	assert.Equal(t, `"Chromium";v="120", "Google Chrome";v="120"`, values.Get("SecCHUA"))
	assert.Equal(t, "x86", values.Get("SecCHUAArch"))
	assert.Equal(t, `"Chromium";v="120.0.6099.129"`, values.Get("SecCHUAFullVersionList"))
	assert.Equal(t, "?0", values.Get("SecCHUAMobile"))
	assert.Equal(t, "Pixel 7", values.Get("SecCHUAModel"))
	assert.Equal(t, "Windows", values.Get("SecCHUAPlatform"))

	// Assert: Verify Sec-Fetch-* headers
	assert.Equal(t, "empty", values.Get("SecFetchDest"))
	assert.Equal(t, "cors", values.Get("SecFetchMode"))
	assert.Equal(t, "same-origin", values.Get("SecFetchSite"))
	assert.Equal(t, "?1", values.Get("SecFetchUser"))

	// Assert: Verify other optional headers
	assert.Equal(t, "1.1 proxy.example.com", values.Get("Via"))
	assert.Equal(t, "webmaster@example.com", values.Get("From"))
	assert.Equal(t, "application/json", values.Get("ContentType"))
	assert.Equal(t, "203.0.113.3", values.Get("TrueClientIP"))

	// Assert: Verify GraphQL fields
	assert.Equal(t, "1", values.Get("GraphQLOperationCount"))
	assert.Equal(t, "getUserProfile", values.Get("GraphQLOperationName"))
	assert.Equal(t, "query", values.Get("GraphQLOperationType"))

	// Assert: Verify encoded string is valid and contains all params
	parsedValues, err := url.ParseQuery(encoded)
	assert.NoError(t, err, "Encoded string should be valid URL query")
	assert.Equal(t, "test-key-456", parsedValues.Get("Key"))
	assert.Equal(t, "getUserProfile", parsedValues.Get("GraphQLOperationName"))
}

func TestProtectionAPIRequestPayload_ToURLValues_GraphQLOperationNamePointer(t *testing.T) {
	t.Run("GraphQLOperationName is nil", func(t *testing.T) {
		// Arrange: Payload with nil GraphQLOperationName
		payload := ProtectionAPIRequestPayload{
			Key:                  "test-key",
			IP:                   "192.168.1.1",
			Request:              "/graphql",
			RequestModuleName:    "Go",
			GraphQLOperationName: nil,
		}

		// Act
		values := payload.ToURLValues()

		// Assert: GraphQLOperationName should not be present
		assert.Empty(t, values.Get("GraphQLOperationName"), "GraphQLOperationName should not be present when nil")
	})

	t.Run("GraphQLOperationName is empty string", func(t *testing.T) {
		// Arrange: Payload with empty string GraphQLOperationName
		emptyName := ""
		payload := ProtectionAPIRequestPayload{
			Key:                  "test-key",
			IP:                   "192.168.1.1",
			Request:              "/graphql",
			RequestModuleName:    "Go",
			GraphQLOperationName: &emptyName,
		}

		// Act
		values := payload.ToURLValues()

		// Assert: GraphQLOperationName should not be present when empty
		assert.Empty(t, values.Get("GraphQLOperationName"), "GraphQLOperationName should not be present when empty string")
	})

	t.Run("GraphQLOperationName has value", func(t *testing.T) {
		// Arrange: Payload with valid GraphQLOperationName
		opName := "getUser"
		payload := ProtectionAPIRequestPayload{
			Key:                  "test-key",
			IP:                   "192.168.1.1",
			Request:              "/graphql",
			RequestModuleName:    "Go",
			GraphQLOperationName: &opName,
		}

		// Act
		values := payload.ToURLValues()

		// Assert: GraphQLOperationName should be present
		assert.Equal(t, "getUser", values.Get("GraphQLOperationName"), "GraphQLOperationName should be present with value")
	})
}

func TestProtectionAPIRequestPayload_ToURLValues_GraphQLOperationType(t *testing.T) {
	testCases := []struct {
		name          string
		operationType OperationType
		expectedValue string
	}{
		{
			name:          "Query operation",
			operationType: Query,
			expectedValue: "query",
		},
		{
			name:          "Mutation operation",
			operationType: Mutation,
			expectedValue: "mutation",
		},
		{
			name:          "Subscription operation",
			operationType: Subscription,
			expectedValue: "subscription",
		},
		{
			name:          "Empty operation type",
			operationType: "",
			expectedValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			payload := ProtectionAPIRequestPayload{
				Key:                  "test-key",
				IP:                   "192.168.1.1",
				Request:              "/graphql",
				RequestModuleName:    "Go",
				GraphQLOperationType: tc.operationType,
			}

			// Act
			values := payload.ToURLValues()

			// Assert
			if tc.expectedValue == "" {
				assert.Empty(t, values.Get("GraphQLOperationType"), "GraphQLOperationType should not be present when empty")
			} else {
				assert.Equal(t, tc.expectedValue, values.Get("GraphQLOperationType"))
			}
		})
	}
}

func TestProtectionAPIRequestPayload_ToURLValues_SpecialCharactersEncoding(t *testing.T) {
	// Arrange: Payload with special characters that need URL encoding
	payload := ProtectionAPIRequestPayload{
		Key:               "test-key",
		IP:                "192.168.1.1",
		Request:           "/api/search?q=hello world&filter=active",
		RequestModuleName: "Go",
		UserAgent:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		Referer:           "https://example.com/path?param=value&other=test",
		Accept:            "text/html,application/xhtml+xml",
	}

	// Act
	encoded := payload.ToURLValues().Encode()

	// Assert: Verify encoding and decoding works correctly
	parsedValues, err := url.ParseQuery(encoded)
	assert.NoError(t, err, "Should be able to parse encoded values")
	assert.Equal(t, "/api/search?q=hello world&filter=active", parsedValues.Get("Request"))
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", parsedValues.Get("UserAgent"))
	assert.Equal(t, "https://example.com/path?param=value&other=test", parsedValues.Get("Referer"))
	assert.Equal(t, "text/html,application/xhtml+xml", parsedValues.Get("Accept"))
}

func TestProtectionAPIRequestPayload_ToURLValues_EmptyOptionalFields(t *testing.T) {
	// Arrange: Payload with some empty optional fields
	payload := ProtectionAPIRequestPayload{
		Key:               "test-key",
		IP:                "192.168.1.1",
		Request:           "/api/endpoint",
		RequestModuleName: "Go",

		// These are explicitly empty
		ModuleVersion: "",
		ServerName:    "",
		UserAgent:     "",
		Port:          "",
	}

	// Act
	values := payload.ToURLValues()

	// Assert: Empty optional fields should not be present in the output
	assert.Empty(t, values.Get("ModuleVersion"))
	assert.Empty(t, values.Get("ServerName"))
	assert.Empty(t, values.Get("UserAgent"))
	assert.Empty(t, values.Get("Port"))

	// Assert: Only 4 parameters should be present (the required ones)
	assert.Len(t, values, 4, "Only required parameters should be present")
}
