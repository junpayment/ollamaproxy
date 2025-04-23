# Ollama-compatible Proxy Server for Lite LLM

This is a Go implementation of an Ollama-compatible proxy server that connects to a Lite LLM backend. The proxy server allows JetBrains AI Assistant to connect to an internal LLM server via Ollama.

## Features

- Implements Ollama-compatible API endpoints:
  - `/api/generate` for text generation
  - `/api/chat` for chat-based interactions
  - `/api/tags` to list available models
- Proxies requests to a Lite LLM backend
- Supports streaming responses
- Handles authentication with API keys
- Configurable base URL for the Lite LLM server
- Comprehensive logging system for monitoring and debugging

## Installation

### Prerequisites

- Go 1.20 or later

### Building from Source

1. Clone the repository:

```bash
git clone <repository-url>
cd ollamaproxy
```

2. Build the application:

```bash
go build -o ollamaproxy
```

## Usage

Run the proxy server with the following command:

```bash
./ollamaproxy --port=11434 --api-key=xxxxxx --base-url=https://xxxxxx
```

### Command-line Arguments

- `--base-url`: (Required) Base URL for the Lite LLM API
- `--api-key`: (Optional) API key for the Lite LLM API
- `--port`: (Optional) Port to run the server on (default: 11434)
- `--host`: (Optional) Host to run the server on (default: 0.0.0.0)

## API Endpoints

### Generate Text

```
POST /api/generate
```

Example request:

```json
{
  "model": "model-name",
  "prompt": "Hello, how are you?",
  "stream": true,
  "system": "You are a helpful assistant."
}
```

### Chat Interaction

```
POST /api/chat
```

Example request:

```json
{
  "model": "model-name",
  "prompt": "Hello, how are you?",
  "stream": true,
  "system": "You are a helpful assistant."
}
```

### List Models

```
GET /api/tags
```

## Logging

The proxy server includes a streamlined logging system focused on access logs only, minimizing noise while providing essential information for monitoring and troubleshooting.

### Access Logs

The server generates access logs in a standard format for every HTTP request. These access logs provide a concise record of all HTTP traffic and include:

- Client IP address
- HTTP method (GET, POST, etc.)
- Request path
- HTTP protocol version
- Response status code
- Response size in bytes
- Request processing duration

Access logs follow this format:
```
ACCESS: [timestamp] client_ip "method path protocol" status_code content_length duration
```

Example:
```
ACCESS: 2023/04/15 12:34:56 127.0.0.1 "POST /api/generate HTTP/1.1" 200 1024 45.6ms
```

### Benefits

This focused logging system is valuable for:
- Monitoring server activity and traffic patterns
- Performance analysis and optimization
- Security auditing and compliance
- Traffic analysis and capacity planning

All logs are output to stdout by default.

## Notes

- This implementation focuses on the core functionality needed to connect JetBrains AI Assistant to your internal LLM server.
- You may need to adjust the request/response format conversion based on the specific implementation of your Lite LLM server.
- Error handling is included but may need to be enhanced for production use.
- Additional Ollama endpoints can be added as needed.
