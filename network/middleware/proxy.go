package middleware

import (
	"iscrie/utils"
	"net/http"
	"net/url"
)

// NewHTTPClientWithProxy creates an HTTP client with optional proxy support.
func NewHTTPClientWithProxy(proxyEnabled bool, proxyURL, proxyUser, proxyPass string) (*http.Client, error) {
	// If proxy is not enabled, return a default HTTP client
	if !proxyEnabled {
		utils.LogDebug("Proxy not enabled. Using default HTTP client.")
		return &http.Client{}, nil
	}

	// Validate proxy URL
	if proxyURL == "" {
		return nil, utils.LogAndReturnError("Proxy is enabled but proxy URL is empty")
	}

	// Parse the proxy URL
	parsedURL, err := url.Parse(proxyURL)
	if err != nil || !parsedURL.IsAbs() {
		return nil, utils.LogAndReturnError("Invalid proxy URL: %w", err)
	}

	// Add user authentication if provided
	if (proxyUser != "" && proxyPass == "") || (proxyUser == "" && proxyPass != "") {
		return nil, utils.LogAndReturnError("proxyUser and proxyPass must both be provided or left empty")
	}
	if proxyUser != "" && proxyPass != "" {
		parsedURL.User = url.UserPassword(proxyUser, proxyPass)
		utils.LogDebug("Proxy authentication configured for user: %s", proxyUser)
	}

	// Configure the HTTP transport with the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedURL),
	}

	utils.LogInfo("Proxy configured successfully. Proxy URL: %s", proxyURL)

	// Return the HTTP client
	return &http.Client{Transport: transport}, nil
}
