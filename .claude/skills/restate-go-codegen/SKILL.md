---
name: restate-go-codegen
description: Guide for Restate Go code generation from Protocol Buffers including type-safe servers and clients. Use when generating Restate code from protobuf definitions or creating type-safe service interfaces.
---

# Restate Go Code Generation

Create type-safe Restate servers and clients from Protocol Buffer service definitions.

## Overview

Restate's Go SDK enables code generation for:
- Type-safe server interfaces
- Type-safe client interfaces
- Automatic serialization (Protocol Buffers or JSON)
- Virtual Objects and Workflows via proto options

## Prerequisites

### Required Tools

**protoc-gen-go-restate**: Restate code generator

```bash
go install github.com/restatedev/sdk-go/protoc-gen-go-restate@latest
```

**protoc**: Protocol Buffer compiler

```bash
# macOS
brew install protobuf

# Linux
apt-get install protobuf-compiler

# Or download from: https://github.com/protocolbuffers/protobuf/releases
```

## Proto Service Definition

### Basic Service

```protobuf
syntax = "proto3";

package example;

option go_package = "github.com/myorg/myproject/gen/example";

// Request message
message GreetRequest {
  string name = 1;
}

// Response message
message GreetResponse {
  string greeting = 1;
}

// Service definition
service Greeter {
  rpc Greet(GreetRequest) returns (GreetResponse);
}
```

### Virtual Object

```protobuf
syntax = "proto3";

package example;

import "dev/restate/ext.proto";

option go_package = "github.com/myorg/myproject/gen/example";

message IncrementRequest {
  int32 delta = 1;
}

message CountResponse {
  int32 count = 1;
}

service Counter {
  option (dev.restate.ext.service_type) = KEYED_SERVICE;

  rpc Increment(IncrementRequest) returns (CountResponse);
  rpc Get(restate.Empty) returns (CountResponse) {
    option (dev.restate.ext.handler_type) = SHARED;
  };
}
```

### Workflow

```protobuf
syntax = "proto3";

package example;

import "dev/restate/ext.proto";

option go_package = "github.com/myorg/myproject/gen/example";

message OrderRequest {
  string order_id = 1;
  repeated string items = 2;
}

message OrderReceipt {
  string receipt_id = 1;
  string status = 2;
}

message StatusRequest {}

message StatusResponse {
  string status = 1;
}

service OrderWorkflow {
  option (dev.restate.ext.service_type) = WORKFLOW;

  // Main workflow entry point
  rpc Run(OrderRequest) returns (OrderReceipt);

  // Query handler (shared)
  rpc GetStatus(StatusRequest) returns (StatusResponse) {
    option (dev.restate.ext.handler_type) = SHARED;
  };
}
```

## Code Generation

### Generate Command

```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-restate_out=. \
  --go-restate_opt=paths=source_relative \
  service.proto
```

### Generated Files

**service.pb.go**: Contains Go type definitions
**service_restate.pb.go**: Contains Restate-specific generated code

## Generated Components

### Server Interface

Generated server interface for implementation:

```go
// Generated interface
type GreeterServer interface {
    Greet(restate.Context, *GreetRequest) (*GreetResponse, error)
}

// Your implementation
type myGreeterServer struct {
    UnimplementedGreeterServer // Backwards compatibility
}

func (s *myGreeterServer) Greet(
    ctx restate.Context,
    req *GreetRequest,
) (*GreetResponse, error) {
    return &GreetResponse{
        Greeting: "Hello, " + req.Name + "!",
    }, nil
}
```

### Client Interface

Generated client for type-safe calls:

```go
// Type-safe service call
client := NewGreeterClient(ctx, "serviceName")
response, err := client.Greet(&GreetRequest{Name: "Alice"})
```

### Serialization Options

Default uses canonical Protobuf JSON encoding. For native Protobuf:

```go
// Client with Protobuf encoding
client := NewGreeterClient(ctx, "serviceName", restate.WithProto())
response, err := client.Greet(&GreetRequest{Name: "Alice"})
```

## Complete Example

### Proto Definition

```protobuf
// example/service.proto
syntax = "proto3";

package example;

import "dev/restate/ext.proto";

option go_package = "github.com/myorg/myproject/gen/example";

message UserRequest {
  string user_id = 1;
}

message UserProfile {
  string user_id = 1;
  string name = 2;
  string email = 3;
}

service UserService {
  rpc GetProfile(UserRequest) returns (UserProfile);
  rpc UpdateProfile(UserProfile) returns (UserProfile);
}
```

### Generate Code

```bash
protoc \
  --go_out=gen \
  --go_opt=paths=source_relative \
  --go-restate_out=gen \
  --go-restate_opt=paths=source_relative \
  example/service.proto
```

### Implement Server

```go
package main

import (
    "github.com/restatedev/sdk-go"
    pb "github.com/myorg/myproject/gen/example"
)

type userServiceServer struct {
    pb.UnimplementedUserServiceServer
}

func (s *userServiceServer) GetProfile(
    ctx restate.Context,
    req *pb.UserRequest,
) (*pb.UserProfile, error) {
    // Fetch from database
    profile, err := restate.Run(ctx, func(ctx restate.RunContext) (*pb.UserProfile, error) {
        return fetchUserProfile(req.UserId)
    })

    return profile, err
}

func (s *userServiceServer) UpdateProfile(
    ctx restate.Context,
    profile *pb.UserProfile,
) (*pb.UserProfile, error) {
    // Update database
    _, err := restate.Run(ctx, func(ctx restate.RunContext) (restate.Void, error) {
        return restate.Void{}, updateUserProfile(profile)
    })

    if err != nil {
        return nil, err
    }

    return profile, nil
}

func main() {
    server := restate.NewServer(restate.WithPort(9080))

    // Register generated service
    pb.RegisterUserServiceServer(server, &userServiceServer{})

    server.Start()
}
```

### Use Client

```go
func callUserService(ctx restate.Context, userID string) (*pb.UserProfile, error) {
    // Create type-safe client
    client := pb.NewUserServiceClient(ctx, "UserService")

    // Make request
    profile, err := client.GetProfile(&pb.UserRequest{
        UserId: userID,
    })

    return profile, err
}
```

## Virtual Object with Code Generation

### Proto Definition

```protobuf
syntax = "proto3";

package example;

import "dev/restate/ext.proto";

option go_package = "github.com/myorg/myproject/gen/example";

message AddItemRequest {
  string product_id = 1;
  int32 quantity = 2;
}

message CartResponse {
  repeated string items = 1;
  int32 total_items = 2;
}

message Empty {}

service ShoppingCart {
  option (dev.restate.ext.service_type) = KEYED_SERVICE;

  rpc AddItem(AddItemRequest) returns (CartResponse);
  rpc GetItems(Empty) returns (CartResponse) {
    option (dev.restate.ext.handler_type) = SHARED;
  };
  rpc Clear(Empty) returns (Empty);
}
```

### Implementation

```go
type shoppingCartServer struct {
    pb.UnimplementedShoppingCartServer
}

func (s *shoppingCartServer) AddItem(
    ctx restate.ObjectContext,
    req *pb.AddItemRequest,
) (*pb.CartResponse, error) {
    // Get current items
    items, _ := restate.Get[[]string](ctx, "items")

    // Add new item
    for i := 0; i < int(req.Quantity); i++ {
        items = append(items, req.ProductId)
    }

    // Update state
    restate.Set(ctx, "items", items)

    return &pb.CartResponse{
        Items:      items,
        TotalItems: int32(len(items)),
    }, nil
}

func (s *shoppingCartServer) GetItems(
    ctx restate.ObjectSharedContext,
) (*pb.CartResponse, error) {
    items, _ := restate.Get[[]string](ctx, "items")

    return &pb.CartResponse{
        Items:      items,
        TotalItems: int32(len(items)),
    }, nil
}

func (s *shoppingCartServer) Clear(
    ctx restate.ObjectContext,
    req *pb.Empty,
) (*pb.Empty, error) {
    restate.ClearAll(ctx)
    return &pb.Empty{}, nil
}
```

### Client Usage

```go
// Create Virtual Object client with key
client := pb.NewShoppingCartClient(ctx, "ShoppingCart", "user-123")

// Add item
response, err := client.AddItem(&pb.AddItemRequest{
    ProductId: "product-456",
    Quantity:  2,
})
```

## Using Buf

Buf simplifies proto management and code generation.

### Install Buf

```bash
# macOS
brew install bufbuild/buf/buf

# Or use binary releases
```

### buf.gen.yaml

```yaml
version: v1
managed:
  enabled: true
plugins:
  - plugin: go
    out: gen
    opt:
      - paths=source_relative
  - plugin: go-restate
    out: gen
    opt:
      - paths=source_relative
```

### buf.yaml

```yaml
version: v1
deps:
  - buf.build/restatedev/proto
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

### Generate with Buf

```bash
buf generate
```

## Service Type Options

### Basic Service (Default)

```protobuf
service MyService {
  rpc Handler(Request) returns (Response);
}
```

### Virtual Object (Keyed Service)

```protobuf
service MyObject {
  option (dev.restate.ext.service_type) = KEYED_SERVICE;

  rpc ExclusiveHandler(Request) returns (Response);
  rpc SharedHandler(Request) returns (Response) {
    option (dev.restate.ext.handler_type) = SHARED;
  };
}
```

### Workflow

```protobuf
service MyWorkflow {
  option (dev.restate.ext.service_type) = WORKFLOW;

  rpc Run(Input) returns (Output);
  rpc Query(QueryRequest) returns (QueryResponse) {
    option (dev.restate.ext.handler_type) = SHARED;
  };
}
```

## Best Practices

1. **Version Proto Files**: Use semantic versioning for proto definitions
2. **Backwards Compatibility**: Use `UnimplementedXServer` for future-proofing
3. **Organized Structure**: Keep proto files in dedicated directory
4. **Use Buf**: Simplifies dependency management and generation
5. **Proto Imports**: Import Restate extensions for service types
6. **Type Safety**: Leverage generated types for compile-time safety
7. **Documentation**: Add comments to proto definitions (appear in generated code)
8. **Consistent Naming**: Follow protobuf naming conventions

## Protobuf vs JSON Serialization

### JSON (Default)

```go
// Canonical Protobuf JSON encoding
client := NewMyServiceClient(ctx, "MyService")
```

**Advantages**:
- Human-readable
- Debug-friendly
- Language-agnostic

### Protobuf Binary

```go
// Native Protobuf binary encoding
client := NewMyServiceClient(ctx, "MyService", restate.WithProto())
```

**Advantages**:
- Smaller payload size
- Faster serialization
- Better performance

## References

- Official Docs: https://docs.restate.dev/develop/go/code-generation
- Protocol Buffers: https://protobuf.dev
- Buf: https://buf.build
- Restate Proto: https://buf.build/restatedev/proto
- Services: See restate-go-services skill
- Client SDK: See restate-go-client skill
