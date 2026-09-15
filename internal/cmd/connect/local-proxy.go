package connect

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/api/project"
)

// RequestDetails represents the structure of the incoming request payload
type RequestDetails struct {
	RequestID string            `json:"request_id"` // Unique ID for the request
	Method    string            `json:"method"`     // HTTP method (e.g., GET, POST)
	Domain    string            `json:"domain"`     // Domain of the URL
	Path      string            `json:"path"`       // Request path (e.g., /foo)
	Headers   map[string]string `json:"headers"`    // Request headers
	Query     map[string]string `json:"query"`      // Query parameters
	Body      string            `json:"body"`       // Request body
	Port      string            `json:"port"`       // Optional port (e.g., "8080")
	IsSecure  bool              `json:"isSecure"`   // To convert to secure network call
}

// ResponseDetails represents the structure of the response to be returned
type ResponseDetails struct {
	RequestID  string            `json:"request_id"`
	StatusCode int               `json:"status_code"` // HTTP status code (e.g., 200, 404)
	Headers    map[string]string `json:"headers"`     // Response headers
	Body       string            `json:"body"`        // Response body
}

// Start a local proxy server and return the port number it's listening on
func StartLocalProxy(selectedOrg *api.Organization, restEndpoint *project.Endpoint, selectedEnv *project.ProjectEnvironment, secureHosts *map[string]bool) (int, error) {
	// Start a local server on a random port
	listener, err := net.Listen("tcp", ":0") // Listen on a random available port
	if err != nil {
		return 0, fmt.Errorf("failed to start local server: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	// Handle incoming requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Forward the request to the server-agent
		forwardRequestToServerAgent(w, r, selectedOrg, restEndpoint, selectedEnv, secureHosts)
	})

	// Start the server in the background
	go func() {
		if err := http.Serve(listener, nil); err != nil {
			log.Fatal("Local proxy server error:", err)
		}
	}()

	return port, nil
}

// Forward incoming requests to the server-agent
func forwardRequestToServerAgent(w http.ResponseWriter, r *http.Request, selectedOrg *api.Organization, restEndpoint *project.Endpoint, selectedEnv *project.ProjectEnvironment, secureHosts *map[string]bool) {
	// Capture the request details
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Convert headers to a map[string]string
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0] // Use the first value for each header
	}

	// Convert query parameters to a map[string]string
	query := make(map[string]string)
	for key, values := range r.URL.Query() {
		query[key] = values[0] // Use the first value for each query parameter
	}

	domain, _ := getDomain(r.URL.String())
	if domain == "" {
		domain = r.Host
	}

	// Extract the port from r.URL.Host
	var requestPort string
	_, requestPort, err = net.SplitHostPort(r.URL.Host)
	if err != nil {
		// If there's no port in the URL, use the default port (empty string)
		requestPort = ""
	}
	isSecure := (*secureHosts)[domain]

	// Create a RequestDetails struct
	reqDetails := RequestDetails{
		Domain:   domain,
		Method:   r.Method,
		Path:     r.URL.Path,
		Headers:  headers,
		Query:    query,
		Body:     string(body),
		Port:     requestPort, // Include the port from the request URL
		IsSecure: isSecure,
	}

	// Marshal the RequestDetails struct to JSON
	reqDetailsJSON, err := json.Marshal(reqDetails)
	if err != nil {
		http.Error(w, "failed to marshal request details", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/handle-client-req", restEndpoint.PublicURL), bytes.NewBuffer(reqDetailsJSON))
	if err != nil {
		log.Println("integration-platform: Failed to generate new request for /handle-client-req:", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	apiKey, err := getProxyAgentRestEpKey(selectedOrg, restEndpoint, selectedEnv)
	if err != nil {
		log.Println("integration-platform: Failed to retrieve API key for bridge REST endpoint:", err)
		return
	}
	apiKeyHeader := "Api-Key"
	if os.Getenv("WSO2IP_ENV") == "dev" {
		apiKeyHeader = "Test-Key"
	}
	req.Header.Add(apiKeyHeader, apiKey.Apikey)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "failed to forward request to bridge component", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body from the /handle-client-req endpoint
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response body from bridge component", http.StatusInternalServerError)
		return
	}

	// Parse the response body into a ResponseDetails struct
	var responseDetails ResponseDetails
	if err := json.Unmarshal(respBody, &responseDetails); err != nil {
		http.Error(w, "failed to parse response details from bridge component", http.StatusInternalServerError)
		return
	}

	// Copy the response headers
	for key, value := range responseDetails.Headers {
		w.Header().Set(key, value)
	}

	// Copy the response status code
	w.WriteHeader(responseDetails.StatusCode)

	// Copy the response body
	w.Write([]byte(responseDetails.Body))
}

func ConnectToServerAgent(userId string, componentHandle string, selectedOrg *api.Organization, wsEndpoint *project.Endpoint, selectedEnv *project.ProjectEnvironment, componentPort int, port int) error {
	// Connect to the WebSocket endpoint
	headers := http.Header{}
	apiKey, err := getProxyAgentWSEpKey(selectedOrg, wsEndpoint, selectedEnv)
	if err != nil {
		log.Println("integration-platform: Failed to retrieve API key for bridge REST endpoint:", err)
		return err
	}
	headers.Set("Sec-WebSocket-Protocol", fmt.Sprintf("ip-test-key, %s", apiKey.Apikey))
	checkRepoAuthSpinner := utils.CreateSpinner(i18n.T(" Generating preview URL..."), "") // TODO: may fail the first time. Need a retry mechanism if fails
	checkRepoAuthSpinner.Start()
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/%s/%s", wsEndpoint.PublicURL, userId, componentHandle), headers)
	checkRepoAuthSpinner.Stop()
	if err != nil {
		return fmt.Errorf("failed to connect to bridge component: %v", err)
	}

	// Start listening for WebSocket messages in the background
	go func() {
		// TODO: if the connection gets disrupted, need to try and reconnect
		defer conn.Close()
		for {
			var reqDetails RequestDetails
			err := conn.ReadJSON(&reqDetails)
			if err != nil {
				log.Println("Integration Platform: Failed to parse request details:", err)
				return
			}

			// Construct the URL with query parameters
			urlStr := fmt.Sprintf("http://localhost:%d%s", componentPort, reqDetails.Path)

			// Create a new HTTP request based on the method
			req, err := http.NewRequest(reqDetails.Method, urlStr, bytes.NewBuffer([]byte(reqDetails.Body)))
			if err != nil {
				log.Println("Integration Platform: Failed to create request:", err)
				continue
			}

			// Set headers from the request details
			for key, value := range reqDetails.Headers {
				req.Header.Set(key, value)
			}

			// Send the request to the local server
			localClient := &http.Client{Timeout: 30 * time.Second}
			localResp, err := localClient.Do(req)
			if err != nil {
				log.Println("Integration Platform: Local request error:", err)
				continue
			}
			defer localResp.Body.Close()

			bodyStr, err := readResponseBody(localResp)
			if err != nil {
				log.Println("Integration Platform: Failed to read local server response:", err)
				continue
			}

			// Prepare the response details
			responseDetails := ResponseDetails{
				RequestID:  reqDetails.RequestID,
				StatusCode: localResp.StatusCode,
				Headers:    make(map[string]string),
				Body:       bodyStr,
			}

			// Copy response headers
			for key, values := range localResp.Header {
				if key != "Content-Encoding" && key != "Content-Length" {
					responseDetails.Headers[key] = values[0]
				}
			}

			// Send the response back to the server
			err = conn.WriteJSON(responseDetails)
			if err != nil {
				log.Println("integration-platform: Failed to send response back to bridge component:", err)
			}
		}
	}()

	// Allow the caller to continue execution
	return nil
}

func readResponseBody(resp *http.Response) (string, error) {
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host, _, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		// If there's no port, return the host as is
		return parsedURL.Host, nil
	}

	return host, nil
}
