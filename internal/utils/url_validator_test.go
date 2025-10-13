package utils

import (
	"errors"
	"net/http"
	"testing"
)

type mockRoundTripper struct {
	responses map[string]*http.Response
	errors    map[string]error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	url := req.URL.String()
	if err, ok := m.errors[url]; ok {
		return nil, err
	}
	if resp, ok := m.responses[url]; ok {
		// Ensure we reuse the request body semantics expected by http.Client
		copy := *resp
		copy.Request = req
		return &copy, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     http.StatusText(http.StatusNotFound),
		Header:     make(http.Header),
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

// TestURLValidation tests the URL validation with problematic URLs from AI responses
func TestURLValidation(t *testing.T) {
	validator := NewURLValidator()

	// Inject mock HTTP client to avoid external network dependencies
	validator.client = &http.Client{
		Transport: &mockRoundTripper{
			responses: map[string]*http.Response{
				"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665515": {
					StatusCode: http.StatusNotFound,
					Status:     http.StatusText(http.StatusNotFound),
					Header:     make(http.Header),
					Body:       http.NoBody,
				},
				"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665523": {
					StatusCode: http.StatusGone,
					Status:     http.StatusText(http.StatusGone),
					Header:     make(http.Header),
					Body:       http.NoBody,
				},
				"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665533": {
					StatusCode: http.StatusForbidden,
					Status:     http.StatusText(http.StatusForbidden),
					Header:     make(http.Header),
					Body:       http.NoBody,
				},
				"https://httpbin.org/image/jpeg": func() *http.Response {
					header := make(http.Header)
					header.Set("Content-Type", "image/jpeg")
					return &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Header:     header,
						Body:       http.NoBody,
					}
				}(),
			},
			errors: map[string]error{
				"https://invalid-url-that-does-not-exist.com/image.jpg": errors.New("dial tcp: no such host"),
			},
		},
	}

	// Test URLs from the user's issue - these should fail validation
	testCases := []struct {
		url      string
		expected bool
		desc     string
	}{
		{
			url:      "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665515",
			expected: false,
			desc:     "Problematic AI-generated URL 1 (should fail)",
		},
		{
			url:      "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665523",
			expected: false,
			desc:     "Problematic AI-generated URL 2 (should fail)",
		},
		{
			url:      "https://chatbot.growrvsb.com/public/images/chatgpt/23141741665533",
			expected: false,
			desc:     "Problematic AI-generated URL 3 (should fail)",
		},
		{
			url:      "https://httpbin.org/image/jpeg",
			expected: true,
			desc:     "Valid image URL (should pass)",
		},
		{
			url:      "https://invalid-url-that-does-not-exist.com/image.jpg",
			expected: false,
			desc:     "Invalid URL (should fail)",
		},
		{
			url:      "not-a-url",
			expected: false,
			desc:     "Invalid format (should fail)",
		},
	}

	t.Log("🧪 Testing URL Validation for AI Response Sanitization")
	t.Log("============================================================")

	for i, tc := range testCases {
		t.Logf("\n%d. %s", i+1, tc.desc)
		t.Logf("   URL: %s", tc.url)

		isValid, mediaType, err := validator.ValidateMediaURL(tc.url)

		if isValid {
			t.Logf("   ✅ VALID - Media Type: %s", mediaType)
		} else {
			t.Logf("   ❌ INVALID - Error: %v", err)
			t.Logf("   📝 Fallback message would be sent instead")
		}

		// Verify the result matches expectation
		if isValid != tc.expected {
			t.Errorf("Expected %v but got %v for URL: %s", tc.expected, isValid, tc.url)
		}
	}

	t.Log("\n============================================================")
	t.Log("🎯 URL validation test completed!")
	t.Log("💡 Invalid URLs will now trigger fallback text messages instead of broken media.")
}

// TestQuickValidation tests the quick validation method
func TestQuickValidation(t *testing.T) {
	validator := NewURLValidator()

	// Test quick validation with problematic URLs
	problemURLs := []string{
		"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665515",
		"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665523",
		"https://chatbot.growrvsb.com/public/images/chatgpt/23141741665533",
	}

	for _, url := range problemURLs {
		isValid := validator.ValidateURLQuick(url)
		t.Logf("Quick validation for %s: %v", url, isValid)

		// These URLs should pass quick validation (format is correct)
		// but will fail full validation due to 404 errors
		if !isValid {
			t.Errorf("URL %s failed quick validation but has correct format", url)
		}
	}
}
