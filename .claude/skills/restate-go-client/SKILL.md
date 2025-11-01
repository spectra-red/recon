---
name: restate-go-client
description: Guide for Restate Go SDK client for invoking handlers from external applications. Use when calling Restate services from outside handlers, implementing client applications, or integrating with Restate from non-Restate code.
---

# Restate Go SDK Client

Invoke Restate handlers from external applications using the Go SDK client.

## Client Creation

Connect to a Restate instance from external code (not within handlers).

### Basic Setup

```go
import "github.com/restatedev/sdk-go/ingress"

// Create client
restateClient := restateingress.NewClient("http://localhost:8080")
```

### Environment-Based Configuration

```go
import (
    "os"
    "github.com/restatedev/sdk-go/ingress"
)

func getRestateClient() *restateingress.Client {
    restateURL := os.Getenv("RESTATE_URL")
    if restateURL == "" {
        restateURL = "http://localhost:8080"
    }
    return restateingress.NewClient(restateURL)
}
```

## Invocation Methods

### Request-Response Invocations

Wait for handler response.

#### Service Call

```go
type Input struct {
    Message string
}

type Output struct {
    Result string
}

// Call service handler
response, err := restateingress.Service[Input, Output](
    client,
    "MyService",      // Service name
    "ProcessMessage", // Handler name
).Request(context.Background(), Input{Message: "hello"})

if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Result)
```

#### Virtual Object Call

```go
// Call Virtual Object handler
response, err := restateingress.Object[Input, Output](
    client,
    "Counter",    // Object name
    "user-123",   // Object key
    "Increment",  // Handler name
).Request(context.Background(), Input{Delta: 1})
```

#### Workflow Call

```go
// Start workflow
receipt, err := restateingress.Workflow[OrderInput, OrderReceipt](
    client,
    "OrderWorkflow", // Workflow name
    "order-456",     // Workflow ID
    "Run",           // Handler name (must be "Run" for workflows)
).Request(context.Background(), orderInput)
```

### One-Way Invocations

Fire-and-forget messaging (doesn't wait for response).

#### Service Send

```go
err := restateingress.ServiceSend(
    client,
    "NotificationService",
    "SendEmail",
).Send(context.Background(), emailData)
```

#### Virtual Object Send

```go
err := restateingress.ObjectSend(
    client,
    "Analytics",
    "user-123",
    "TrackEvent",
).Send(context.Background(), eventData)
```

#### Workflow Send

```go
err := restateingress.WorkflowSend(
    client,
    "OrderWorkflow",
    "order-456",
    "Cancel",
).Send(context.Background(), cancelData)
```

### Delayed Invocations

Schedule execution for a future time.

```go
import "time"

// Delayed service call
err := restateingress.ServiceSend(
    client,
    "ReminderService",
    "SendReminder",
).Send(
    context.Background(),
    reminderData,
    restate.WithDelay(24*time.Hour),
)

// Delayed Virtual Object call
err := restateingress.ObjectSend(
    client,
    "Subscription",
    "user-123",
    "ProcessRenewal",
).Send(
    context.Background(),
    renewalData,
    restate.WithDelay(30*24*time.Hour),
)
```

## Idempotency

Make calls idempotent using idempotency keys.

### With Idempotency Key

```go
// Idempotent service call
response, err := restateingress.Service[PaymentRequest, PaymentResponse](
    client,
    "PaymentService",
    "ProcessPayment",
).Request(
    context.Background(),
    paymentRequest,
    restate.WithIdempotencyKey("payment-order-123"),
)
```

**Guarantees**:
- Same idempotency key returns cached response
- Responses persist for 24 hours
- Prevents duplicate executions

### Idempotency with Delays

```go
err := restateingress.ServiceSend(
    client,
    "EmailService",
    "SendConfirmation",
).Send(
    context.Background(),
    emailData,
    restate.WithIdempotencyKey("email-order-123"),
    restate.WithDelay(1*time.Hour),
)
```

## Result Retrieval

### Attach to Invocation by ID

```go
// Attach to previously started invocation
invocationID := "inv_abc123..."

result, err := restateingress.InvocationById[Response](
    client,
    invocationID,
).Attach(context.Background())

if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

### Attach by Idempotency Key

```go
// For services
result, err := restateingress.ServiceInvocationByIdempotencyKey[Response](
    client,
    "PaymentService",
    "processPayment/my-idempotency-key",
).Attach(context.Background())
```

### Workflow Attachment

```go
// Attach to workflow by ID
result, err := restateingress.WorkflowHandle[Receipt](
    client,
    "OrderWorkflow",
    "order-456",
).Attach(context.Background())

// Peek at workflow result without attaching
result, err := restateingress.WorkflowHandle[Receipt](
    client,
    "OrderWorkflow",
    "order-456",
).Output(context.Background())
```

## Complete Examples

### CLI Application

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"

    "github.com/restatedev/sdk-go/ingress"
)

type GreetRequest struct {
    Name string
}

type GreetResponse struct {
    Greeting string
}

func main() {
    name := flag.String("name", "World", "Name to greet")
    flag.Parse()

    // Create client
    client := restateingress.NewClient("http://localhost:8080")

    // Call service
    response, err := restateingress.Service[GreetRequest, GreetResponse](
        client,
        "GreeterService",
        "Greet",
    ).Request(context.Background(), GreetRequest{Name: *name})

    if err != nil {
        log.Fatalf("Failed to greet: %v", err)
    }

    fmt.Println(response.Greeting)
}
```

### Web Application Handler

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/restatedev/sdk-go/ingress"
)

var restateClient = restateingress.NewClient("http://localhost:8080")

type CreateOrderRequest struct {
    UserID string   `json:"user_id"`
    Items  []string `json:"items"`
}

type OrderResponse struct {
    OrderID string `json:"order_id"`
    Status  string `json:"status"`
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Call Restate workflow
    response, err := restateingress.Workflow[CreateOrderRequest, OrderResponse](
        restateClient,
        "OrderWorkflow",
        generateOrderID(),
        "Run",
    ).Request(r.Context(), req)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    http.HandleFunc("/orders", createOrderHandler)
    http.ListenAndServe(":8000", nil)
}
```

### Background Job Scheduler

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/restatedev/sdk-go/ingress"
)

type TaskData struct {
    TaskID      string
    Description string
}

func scheduleRecurringTask(client *restateingress.Client, taskID string) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        err := restateingress.ServiceSend(
            client,
            "TaskProcessor",
            "ProcessTask",
        ).Send(context.Background(), TaskData{
            TaskID:      taskID,
            Description: "Daily backup",
        })

        if err != nil {
            log.Printf("Failed to schedule task: %v", err)
        } else {
            log.Printf("Task %s scheduled successfully", taskID)
        }
    }
}

func main() {
    client := restateingress.NewClient("http://localhost:8080")
    scheduleRecurringTask(client, "daily-backup")
}
```

### Webhook Handler

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/restatedev/sdk-go/ingress"
)

var restateClient = restateingress.NewClient("http://localhost:8080")

type WebhookPayload struct {
    EventType string                 `json:"event_type"`
    Data      map[string]interface{} `json:"data"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    var payload WebhookPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Send to Restate for processing (fire-and-forget)
    err := restateingress.ServiceSend(
        restateClient,
        "WebhookProcessor",
        "ProcessEvent",
    ).Send(r.Context(), payload)

    if err != nil {
        http.Error(w, "Failed to queue webhook", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusAccepted)
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    http.ListenAndServe(":8000", nil)
}
```

### Idempotent Payment Processing

```go
package main

import (
    "context"
    "fmt"

    "github.com/restatedev/sdk-go/ingress"
)

type PaymentRequest struct {
    OrderID string
    Amount  int
    CardID  string
}

type PaymentResponse struct {
    TransactionID string
    Status        string
}

func processPayment(
    client *restateingress.Client,
    orderID string,
    amount int,
    cardID string,
) (*PaymentResponse, error) {
    // Use order ID as idempotency key
    idempotencyKey := fmt.Sprintf("payment-%s", orderID)

    response, err := restateingress.Service[PaymentRequest, PaymentResponse](
        client,
        "PaymentService",
        "Charge",
    ).Request(
        context.Background(),
        PaymentRequest{
            OrderID: orderID,
            Amount:  amount,
            CardID:  cardID,
        },
        restate.WithIdempotencyKey(idempotencyKey),
    )

    return &response, err
}

func main() {
    client := restateingress.NewClient("http://localhost:8080")

    // This call is idempotent - safe to retry
    response, err := processPayment(client, "order-123", 1000, "card-456")
    if err != nil {
        // Can safely retry - won't double-charge
        response, err = processPayment(client, "order-123", 1000, "card-456")
    }

    fmt.Printf("Payment status: %s\n", response.Status)
}
```

### Workflow Status Polling

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/restatedev/sdk-go/ingress"
)

type WorkflowStatus struct {
    Status    string
    Progress  int
    Completed bool
}

func waitForWorkflowCompletion(
    client *restateingress.Client,
    workflowID string,
    timeout time.Duration,
) (*WorkflowStatus, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("workflow did not complete within timeout")

        case <-ticker.C:
            // Query workflow status
            status, err := restateingress.Workflow[any, WorkflowStatus](
                client,
                "OrderWorkflow",
                workflowID,
                "GetStatus",
            ).Request(context.Background(), nil)

            if err != nil {
                return nil, err
            }

            if status.Completed {
                return &status, nil
            }

            fmt.Printf("Workflow progress: %d%%\n", status.Progress)
        }
    }
}

func main() {
    client := restateingress.NewClient("http://localhost:8080")

    status, err := waitForWorkflowCompletion(client, "order-123", 5*time.Minute)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Workflow completed with status: %s\n", status.Status)
}
```

## Best Practices

1. **Reuse Clients**: Create client once, reuse across requests
2. **Use Context**: Pass appropriate context for cancellation and timeouts
3. **Idempotency Keys**: Use for critical operations to prevent duplicates
4. **Error Handling**: Always check and handle errors appropriately
5. **Type Safety**: Leverage generic types for compile-time safety
6. **Connection Pooling**: Client handles connection pooling internally
7. **Environment Config**: Use environment variables for Restate URL
8. **Graceful Degradation**: Handle Restate unavailability gracefully

## Error Handling

```go
response, err := restateingress.Service[Input, Output](
    client,
    "MyService",
    "Handler",
).Request(context.Background(), input)

if err != nil {
    // Check error type
    if errors.Is(err, context.DeadlineExceeded) {
        // Timeout occurred
        log.Println("Request timed out")
    } else if errors.Is(err, context.Canceled) {
        // Context was cancelled
        log.Println("Request cancelled")
    } else {
        // Other errors (network, Restate errors, handler errors)
        log.Printf("Request failed: %v", err)
    }
    return err
}
```

## Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := restateingress.Service[Input, Output](
    client,
    "MyService",
    "Handler",
).Request(ctx, input)
```

## References

- Official Docs: https://docs.restate.dev/services/invocation/clients/go-sdk
- Service Communication: See restate-go-service-communication skill
- Invocation Management: See restate-invocation-management skill
- Services: See restate-go-services skill
