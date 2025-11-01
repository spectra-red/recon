---
name: restate-go-serving
description: Guide for serving Restate Go services including HTTP server setup, Lambda deployment, and bidirectional communication. Use when deploying services, configuring servers, or integrating with serverless platforms.
---

# Restate Go Serving

Deploy Restate services as HTTP servers or serverless functions.

## Core Concept

The Restate SDK operates as an HTTP handler, deployable as:
- Standalone HTTP/2 server
- Custom HTTP server
- AWS Lambda function
- Other serverless platforms

## HTTP/2 Server (Standard Deployment)

Create a basic HTTP/2 server listening on a specific port.

### Basic Setup

```go
package main

import (
    "github.com/restatedev/sdk-go"
)

type GreeterService struct{}

func (g *GreeterService) Greet(ctx restate.Context, name string) (string, error) {
    return "Hello, " + name + "!", nil
}

func main() {
    // Create server
    server := restate.NewServer(restate.WithPort(9080))

    // Register services using reflection
    server.Bind(restate.Reflect(&GreeterService{}))

    // Start listening
    server.Start()
}
```

### Multiple Service Registration

```go
func main() {
    server := restate.NewServer(restate.WithPort(9080))

    // Register multiple service types
    server.Bind(restate.Reflect(&GreeterService{}))      // Service
    server.Bind(restate.Reflect(&Counter{}))             // Virtual Object
    server.Bind(restate.Reflect(&OrderWorkflow{}))       // Workflow

    server.Start()
}
```

### Configuration Options

```go
server := restate.NewServer(
    restate.WithPort(9080),              // Listen port
    restate.WithLogger(customLogger),    // Custom slog handler
)
```

## Custom HTTP Server

Get the handler for integration with custom HTTP servers.

### HTTP/2 with Custom Server

```go
func main() {
    // Create Restate server
    server := restate.NewServer()
    server.Bind(restate.Reflect(&MyService{}))

    // Get HTTP handler
    handler := server.Handler()

    // Use with custom server
    http.Handle("/", handler)
    http.ListenAndServe(":9080", nil)
}
```

### HTTP/1.1 Support

```go
func main() {
    server := restate.NewServer()
    server.Bind(restate.Reflect(&MyService{}))

    // Disable bidirectional communication for HTTP/1.1
    server.Bidirectional(false)

    handler := server.Handler()

    // HTTP/1.1 server
    http.ListenAndServe(":9080", handler)
}
```

**Important**: Use `--use-http1.1` CLI flag when registering with Restate:

```bash
restate deployments register http://localhost:9080 --use-http1.1
```

## AWS Lambda Deployment

Deploy Restate services as Lambda functions.

### Lambda Handler

```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/restatedev/sdk-go"
)

type MyService struct{}

func (s *MyService) Handle(ctx restate.Context, input string) (string, error) {
    return "Processed: " + input, nil
}

func main() {
    // Create server
    server := restate.NewServer()

    // CRITICAL: Disable bidirectional communication for Lambda
    server.Bidirectional(false)

    // Register service
    server.Bind(restate.Reflect(&MyService{}))

    // Get Lambda handler
    lambdaHandler := server.LambdaHandler()

    // Start Lambda runtime
    lambda.Start(lambdaHandler)
}
```

### Why Disable Bidirectional Communication?

Lambda doesn't support bidirectional communication. Enabling it causes handler deadlocks.

**Required**:
```go
server.Bidirectional(false)
```

### Lambda Deployment Package

```bash
# Build for Lambda
GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# Create deployment package
zip function.zip bootstrap

# Deploy with AWS CLI
aws lambda create-function \
  --function-name restate-service \
  --runtime provided.al2 \
  --handler bootstrap \
  --zip-file fileb://function.zip \
  --role arn:aws:iam::ACCOUNT:role/lambda-role
```

## Security: Identity Verification

Validate requests using public keys.

### Enable Identity Verification

```go
publicKey := "publickeyv1_..."  // Your public key

server := restate.NewServer(restate.WithPort(9080))
server.Bind(restate.Reflect(&MyService{}))

// Enable identity verification
server.WithIdentityV1(publicKey)

server.Start()
```

### Multiple Public Keys

```go
publicKeys := []string{
    "publickeyv1_key1...",
    "publickeyv1_key2...",
}

for _, key := range publicKeys {
    server.WithIdentityV1(key)
}
```

## Service Consistency Across Deployments

Service implementations remain consistent regardless of deployment environment:

```go
type MyService struct{}

func (s *MyService) Process(ctx restate.Context, data Data) (Result, error) {
    // Same code works in:
    // - HTTP/2 server
    // - HTTP/1.1 server
    // - AWS Lambda
    // - Other serverless platforms
    return processData(data), nil
}
```

## Complete Examples

### Production HTTP/2 Server

```go
package main

import (
    "log/slog"
    "os"

    "github.com/restatedev/sdk-go"
)

type OrderService struct{}
type PaymentService struct{}
type InventoryObject struct{}

func main() {
    // Custom logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))

    // Create server with configuration
    server := restate.NewServer(
        restate.WithPort(9080),
        restate.WithLogger(logger),
    )

    // Register services
    server.Bind(restate.Reflect(&OrderService{}))
    server.Bind(restate.Reflect(&PaymentService{}))
    server.Bind(restate.Reflect(&InventoryObject{}))

    // Enable security if public key available
    if publicKey := os.Getenv("RESTATE_PUBLIC_KEY"); publicKey != "" {
        server.WithIdentityV1(publicKey)
    }

    // Start server
    log.Println("Starting Restate server on :9080")
    server.Start()
}
```

### Lambda with Environment Configuration

```go
package main

import (
    "log/slog"
    "os"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/restatedev/sdk-go"
)

type APIService struct{}

func main() {
    // Custom JSON logger for CloudWatch
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    server := restate.NewServer(restate.WithLogger(logger))

    // Required for Lambda
    server.Bidirectional(false)

    server.Bind(restate.Reflect(&APIService{}))

    // Security
    if publicKey := os.Getenv("RESTATE_PUBLIC_KEY"); publicKey != "" {
        server.WithIdentityV1(publicKey)
    }

    // Start Lambda
    lambda.Start(server.LambdaHandler())
}
```

### Multi-Environment Setup

```go
package main

import (
    "os"

    "github.com/restatedev/sdk-go"
)

func main() {
    server := restate.NewServer()
    server.Bind(restate.Reflect(&MyService{}))

    // Detect environment
    if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
        // Lambda environment
        server.Bidirectional(false)
        lambda.Start(server.LambdaHandler())
    } else {
        // Local/container environment
        port := os.Getenv("PORT")
        if port == "" {
            port = "9080"
        }
        server := restate.NewServer(restate.WithPort(parsePort(port)))
        server.Start()
    }
}
```

## Deployment Registration

After starting your service, register it with Restate server.

### HTTP/2 Registration

```bash
restate deployments register http://localhost:9080
```

### HTTP/1.1 Registration

```bash
restate deployments register http://localhost:9080 --use-http1.1
```

### Lambda Registration

```bash
# Get Lambda function URL
FUNCTION_URL=$(aws lambda get-function-url-config \
  --function-name restate-service \
  --query 'FunctionUrl' \
  --output text)

# Register with Restate
restate deployments register $FUNCTION_URL
```

## Server Configuration Options

### Port Configuration

```go
// Default port
server := restate.NewServer() // Uses 9080

// Custom port
server := restate.NewServer(restate.WithPort(8080))

// From environment
port := os.Getenv("PORT")
if port == "" {
    port = "9080"
}
server := restate.NewServer(restate.WithPort(mustParsePort(port)))
```

### Logger Configuration

```go
// JSON logger
jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
server := restate.NewServer(restate.WithLogger(jsonHandler))

// Pretty logger (development)
prettyHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
server := restate.NewServer(restate.WithLogger(prettyHandler))

// Custom logger
customLogger := slog.New(yourCustomHandler)
server := restate.NewServer(restate.WithLogger(customLogger))
```

## Health Checks

Add health check endpoints for production deployments:

```go
func main() {
    server := restate.NewServer(restate.WithPort(9080))
    server.Bind(restate.Reflect(&MyService{}))

    // Restate handler
    restateHandler := server.Handler()

    // Health check endpoint
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // Restate handler for main path
    http.Handle("/", restateHandler)

    // Start server
    http.ListenAndServe(":9080", nil)
}
```

## Best Practices

1. **Use HTTP/2**: Default for best performance and bidirectional support
2. **Disable Bidirectional for Lambda**: Required to prevent deadlocks
3. **Enable Security**: Use identity verification in production
4. **Configure Logging**: Use structured JSON logging for production
5. **Health Checks**: Add health endpoints for monitoring
6. **Environment Variables**: Use env vars for configuration
7. **Graceful Shutdown**: Implement shutdown handlers for containers
8. **Port Configuration**: Make ports configurable via environment

## Troubleshooting

### Lambda Deadlock

**Problem**: Lambda function times out or hangs

**Solution**: Ensure bidirectional communication is disabled
```go
server.Bidirectional(false)
```

### HTTP/1.1 Registration Fails

**Problem**: Service registered but handlers don't work

**Solution**: Use `--use-http1.1` flag
```bash
restate deployments register http://localhost:9080 --use-http1.1
```

### Security Verification Fails

**Problem**: Requests rejected with 401

**Solution**: Ensure public key matches Restate server configuration
```go
server.WithIdentityV1(correctPublicKey)
```

## References

- Official Docs: https://docs.restate.dev/develop/go/serving
- Service Types: See restate-go-services skill
- Logging: See restate-go-logging skill
- Deployment: See restate-docker-deploy skill
