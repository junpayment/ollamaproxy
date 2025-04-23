package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Configuration for the proxy
type Config struct {
	BaseURL string
	APIKey  string
	Port    int
	Host    string
}

// GenerationRequest represents the Ollama API request format
type GenerationRequest struct {
	Model     string                 `json:"model"`
	Stream    bool                   `json:"stream,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	KeepAlive string                 `json:"keep_alive,omitempty"`
	Messages  []Message              `json:"messages"`
}

// LiteLLMRequest represents the Lite LLM API request format
type LiteLLMRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LiteLLMResponse represents the Lite LLM API response format
type LiteLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta,omitempty"`
	} `json:"choices"`
}

// OllamaResponse represents the Ollama API response format
type OllamaResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// ModelResponse represents the Lite LLM models response
type ModelResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// OllamaModelResponse represents the Ollama models response format
type OllamaModelResponse struct {
	Models []struct {
		Name       string                 `json:"name"`
		Model      string                 `json:"model"`
		ModifiedAt string                 `json:"modified_at"`
		Size       int                    `json:"size"`
		Digest     string                 `json:"digest"`
		Details    map[string]interface{} `json:"details"`
	} `json:"models"`
}

var config Config
var accessLogger *log.Logger

func init() {
	// Initialize access logger with timestamp only
	accessLogger = log.New(os.Stdout, "ACCESS: ", log.Ldate|log.Ltime)
}

// accessLogMiddleware logs HTTP requests in a standard access log format
func accessLogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Create a response writer wrapper to capture the status code
		rwWrapper := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
			contentLength:  0,
		}

		// Process the request
		next(rwWrapper, r)

		// Calculate request duration
		duration := time.Since(startTime)

		// Log the request in a standard access log format
		// Format: client_ip [timestamp] "method path protocol" status content_length duration
		accessLogger.Printf("%s \"%s %s %s\" %d %d %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			rwWrapper.statusCode,
			rwWrapper.contentLength,
			duration,
		)
	}
}

// responseWriterWrapper is a wrapper for http.ResponseWriter that captures the status code and content length
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

// WriteHeader captures the status code
func (rww *responseWriterWrapper) WriteHeader(statusCode int) {
	rww.statusCode = statusCode
	rww.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the content length
func (rww *responseWriterWrapper) Write(b []byte) (int, error) {
	size, err := rww.ResponseWriter.Write(b)
	rww.contentLength += size
	return size, err
}

// Flush flushes the response writer if it implements http.Flusher
func (rww *responseWriterWrapper) Flush() {
	if f, ok := rww.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func main() {
	// Parse command line arguments
	flag.StringVar(&config.BaseURL, "base-url", "", "Base URL for Lite LLM API")
	flag.StringVar(&config.APIKey, "api-key", "", "API key for Lite LLM")
	flag.IntVar(&config.Port, "port", 11434, "Port to run the server on")
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host to run the server on")
	flag.Parse()

	if config.BaseURL == "" {
		log.Fatal("Base URL is required")
	}

	// Set up HTTP routes with access logging middleware
	http.HandleFunc("/", accessLogMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ollama is running"))
	}))
	http.HandleFunc("/api/chat", accessLogMiddleware(handleChat))
	http.HandleFunc("/api/tags", accessLogMiddleware(handleListModels))

	// Start the server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var ollamaReq GenerationRequest
	if err := json.Unmarshal(b, &ollamaReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// オプションを追加
	if ollamaReq.Options == nil {
		ollamaReq.Options = make(map[string]interface{})
	}
	ollamaReq.Options["litellm_settings"] = map[string]interface{}{
		"modify_params": true,
	}

	liteLLMReq := LiteLLMRequest{
		Model:    ollamaReq.Model,
		Messages: ollamaReq.Messages,
		Stream:   ollamaReq.Stream,
	}

	// Prepare the request to Lite LLM
	reqBody, err := json.Marshal(liteLLMReq)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Create a new request to Lite LLM
	liteLLMURL := fmt.Sprintf("%s/chat/completions", config.BaseURL)
	req, err := http.NewRequest("POST", liteLLMURL, bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	if ollamaReq.Stream {
		// Handle streaming response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to connect to Lite LLM", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, string(body), resp.StatusCode)
			return
		}

		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					// Silently handle error
				}
				break
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Handle SSE format
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
			}

			if line == "[DONE]" {
				break
			}

			var liteLLMResp LiteLLMResponse
			if err := json.Unmarshal([]byte(line), &liteLLMResp); err != nil {
				continue
			}

			// Convert to Ollama format
			if len(liteLLMResp.Choices) > 0 {
				content := liteLLMResp.Choices[0].Delta.Content
				if content != "" {
					ollamaResp := OllamaResponse{
						Model:   ollamaReq.Model,
						Message: Message{Content: content, Role: "assistant"},
						Done:    false,
					}
					respBytes, _ := json.Marshal(ollamaResp)
					fmt.Fprintf(w, "%s\n", respBytes)
					flusher.Flush()
				}
			}
		}

		// Send final done message
		ollamaResp := OllamaResponse{
			Model:   ollamaReq.Model,
			Message: Message{Content: "", Role: "assistant"},
			Done:    true,
		}
		respBytes, _ := json.Marshal(ollamaResp)
		fmt.Fprintf(w, "%s\n", respBytes)
		flusher.Flush()
	} else {
		// Handle non-streaming response
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to connect to Lite LLM", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, string(body), resp.StatusCode)
			return
		}

		var liteLLMResp LiteLLMResponse
		if err := json.NewDecoder(resp.Body).Decode(&liteLLMResp); err != nil {
			http.Error(w, "Failed to decode response", http.StatusInternalServerError)
			return
		}

		// Convert to Ollama format
		content := ""
		if len(liteLLMResp.Choices) > 0 {
			content = liteLLMResp.Choices[0].Message.Content
		}

		ollamaResp := OllamaResponse{
			Model:   ollamaReq.Model,
			Message: Message{Content: content, Role: "assistant"},
			Done:    true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaResp)
	}
}

func handleListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a new request to Lite LLM
	liteLLMURL := fmt.Sprintf("%s/models", config.BaseURL)
	req, err := http.NewRequest("GET", liteLLMURL, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+config.APIKey)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to connect to Lite LLM", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)
		return
	}

	var modelResp ModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelResp); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Convert to Ollama format
	ollamaModels := OllamaModelResponse{
		Models: make([]struct {
			Name       string                 `json:"name"`
			Model      string                 `json:"model"`
			ModifiedAt string                 `json:"modified_at"`
			Size       int                    `json:"size"`
			Digest     string                 `json:"digest"`
			Details    map[string]interface{} `json:"details"`
		}, len(modelResp.Data)),
	}

	for i, model := range modelResp.Data {
		ollamaModels.Models[i] = struct {
			Name       string                 `json:"name"`
			Model      string                 `json:"model"`
			ModifiedAt string                 `json:"modified_at"`
			Size       int                    `json:"size"`
			Digest     string                 `json:"digest"`
			Details    map[string]interface{} `json:"details"`
		}{
			Name:       model.ID,
			Model:      model.ID,
			ModifiedAt: "",
			Size:       0,
			Digest:     "",
			Details:    make(map[string]interface{}),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ollamaModels)
}
