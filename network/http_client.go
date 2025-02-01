package network

import (
	"errors"
	"fmt"
	"iscrie/config"
	"iscrie/utils"
	"net/http"
	"net/url"
	"os"
	"time"
)

// AuthType defines the type of authentication to use.
type AuthType string

const (
	BasicAuth  AuthType = "basic"
	BearerAuth AuthType = "bearer"
	HeaderAuth AuthType = "header"
)

type HTTPClient struct {
	Client        *http.Client
	Authenticator *Authenticator
}

// HTTPClientAdapter wraps HTTPClient and implements FileUploader.
type HTTPClientAdapter struct {
	HTTPClient   *HTTPClient
	BaseURL      string
	Repository   string
	ForceReplace bool
}

// NewHTTPClientAdapter creates a new instance of HTTPClientAdapter.
func NewHTTPClientAdapter(httpClient *HTTPClient, baseURL, repository string, forceReplace bool) *HTTPClientAdapter {
	utils.LogDebug("Initializing HTTPClientAdapter for repository: %s", repository)
	return &HTTPClientAdapter{
		HTTPClient:   httpClient,
		BaseURL:      baseURL,
		Repository:   repository,
		ForceReplace: forceReplace,
	}
}

// Authenticator manages authentication for HTTP requests.
type Authenticator struct {
	authType    AuthType
	userToken   string
	passToken   string
	accessToken string
	headerName  string
	headerValue string
}

// NewAuthenticator creates and validates an Authenticator.
func NewAuthenticator(authConfig config.AuthConfig) (*Authenticator, error) {
	auth := &Authenticator{
		authType:    AuthType(authConfig.Type),
		userToken:   authConfig.UserToken,
		passToken:   authConfig.PassToken,
		accessToken: authConfig.AccessToken,
		headerName:  authConfig.HeaderName,
		headerValue: authConfig.HeaderValue,
	}

	switch auth.authType {
	case BasicAuth:
		if auth.userToken == "" || auth.passToken == "" {
			return nil, errors.New("basic authentication requires both userToken and passToken")
		}
	case BearerAuth:
		if auth.accessToken == "" {
			return nil, errors.New("bearer authentication requires accessToken")
		}
	case HeaderAuth:
		if auth.headerName == "" || auth.headerValue == "" {
			return nil, errors.New("header authentication requires headerName and headerValue")
		}
	default:
		return nil, fmt.Errorf("unsupported authentication type: %s", auth.authType)
	}

	utils.LogDebug("Authenticator initialized with type: %s", auth.authType)
	return auth, nil
}

// Apply applies the appropriate authentication headers to an HTTP request.
func (a *Authenticator) Apply(req *http.Request) error {
	switch a.authType {
	case BasicAuth:
		req.SetBasicAuth(a.userToken, a.passToken)
		utils.LogDebug("Applied BasicAuth with user: %s", a.userToken)
	case BearerAuth:
		req.Header.Set("Authorization", "Bearer "+a.accessToken)
		utils.LogDebug("Applied BearerAuth with token.")
	case HeaderAuth:
		req.Header.Set(a.headerName, a.headerValue)
		utils.LogDebug("Applied HeaderAuth: %s: %s", a.headerName, a.headerValue)
	default:
		return fmt.Errorf("unsupported authentication type: %s", a.authType)
	}
	return nil
}

// NewHTTPClient creates an HTTPClient with optional proxy and authentication.
func NewHTTPClient(authConfig config.AuthConfig, proxyConfig config.ProxyConfig) (*HTTPClient, error) {
	authenticator, err := NewAuthenticator(authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authenticator: %w", err)
	}

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			if proxyConfig.Enabled {
				if proxyConfig.Host == "" || proxyConfig.Port == 0 {
					return nil, errors.New("proxy is enabled but host or port is not defined")
				}

				proxyURL := &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%d", proxyConfig.Host, proxyConfig.Port),
				}
				if proxyConfig.Username != "" && proxyConfig.Password != "" {
					proxyURL.User = url.UserPassword(proxyConfig.Username, proxyConfig.Password)
				}
				utils.LogDebug("Proxy configured: %s", proxyURL.String())
				return proxyURL, nil
			}
			return nil, nil
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	utils.LogInfo("HTTP client initialized with timeout: 30s")
	return &HTTPClient{
		Client:        client,
		Authenticator: authenticator,
	}, nil
}

// AddCommonHeaders adds common headers to an HTTP request.
func AddCommonHeaders(req *http.Request, forceReplace bool) {
	req.Header.Set("X-Content-Force-Replace", fmt.Sprintf("%v", forceReplace))
	req.Header.Set("Content-Type", "application/json")
	utils.LogDebug("Common headers added: Force-Replace=%v", forceReplace)
}

// UploadFile handles uploading a file to the specified URL using a PUT request.
func (hc *HTTPClient) UploadFile(url string, filePath string, forceReplace bool) (*http.Response, error) {
	utils.LogInfo("Uploading file: %s to URL: %s", filePath, url)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, utils.LogAndReturnError("Failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return nil, utils.LogAndReturnError("Failed to create PUT request: %w", err)
	}

	AddCommonHeaders(req, forceReplace)

	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, utils.LogAndReturnError("HTTP Request failed: %w", err)
	}

	utils.LogDebug("HTTP Response Status: %d", resp.StatusCode)
	return resp, nil
}

// CreatePutRequest prepares a PUT request for uploading a file.
func (hc *HTTPClientAdapter) CreatePutRequest(urlStr, filePath string) (*http.Request, *os.File, error) {
	utils.LogDebug("Preparing PUT request for URL: %s", urlStr)
	utils.LogDebug("File path: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, utils.LogAndReturnError("Failed to open file '%s': %w", filePath, err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, utils.LogAndReturnError("Failed to retrieve file information for '%s': %w", filePath, err)
	}

	if fileInfo.Size() == 0 {
		file.Close()
		return nil, nil, utils.LogAndReturnError("File '%s' is empty", filePath)
	}

	req, err := http.NewRequest(http.MethodPut, urlStr, file)
	if err != nil {
		file.Close()
		return nil, nil, utils.LogAndReturnError("Failed to create PUT request: %w", err)
	}

	AddCommonHeaders(req, false)
	return req, file, nil
}

// Do executes a generic HTTP request and logs details about it.
func (hc *HTTPClientAdapter) Do(req *http.Request) (*http.Response, error) {
	utils.LogDebug("Executing HTTP request...")

	// Apply authentication
	if hc.HTTPClient.Authenticator != nil {
		if err := hc.HTTPClient.Authenticator.Apply(req); err != nil {
			return nil, utils.LogAndReturnError("Failed to apply authentication: %w", err)
		}
	}

	// Execute the request
	resp, err := hc.HTTPClient.Client.Do(req)

	// Log the response details
	if resp != nil {
		utils.LogDebug("HTTP Response Status: %d", resp.StatusCode)
	}
	if err != nil {
		return nil, utils.LogAndReturnError("HTTP Request failed: %w", err)
	}

	return resp, err
}
