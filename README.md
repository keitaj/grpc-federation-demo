# gRPC Federation Demo

This is a demonstration application showcasing how to use [gRPC Federation](https://github.com/mercari/grpc-federation) to build a BFF (Backend for Frontend) API that aggregates data from multiple microservices.

## Architecture

The demo consists of:
- **3 Backend Services**: User, Product, and Order services
- **1 BFF Service**: Aggregates data from backend services using gRPC Federation
- **Test Client**: Demonstrates API usage

## Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for development only)
- Protocol Buffers compiler (protoc) (for development only)

## Quick Start

### 1. Build and run with Docker Compose

```bash
# Build all Docker images
make build-all

# Start all services
make up-all
```

This will start:
- User Service on port 50051
- Product Service on port 50052
- Order Service on port 50053
- BFF Service on port 50054

### 2. Test the BFF API

Using grpcurl:

```bash
# Test GetUserWithOrders (includes success, USER_NOT_FOUND, USER_ACCOUNT_SUSPENDED)
make grpcurl-user

# Test GetOrderDetails (includes success, ORDER_NOT_FOUND, ORDER_CANCELLED, PRODUCT_NOT_FOUND, PRODUCT_OUT_OF_STOCK)
make grpcurl-order

# Test GetUserDashboard (includes success, USER_NOT_FOUND, USER_ACCOUNT_SUSPENDED)
make grpcurl-dashboard

# Or test all APIs and all error cases at once
make grpcurl-all
```

**Test Cases:**
- **Success cases**: Normal user/order data
- **FailedPrecondition cases**: Various business logic errors (user not found, account suspended, order cancelled, product issues)
- **Error handling**: Demonstrates localized Japanese error messages with detailed error information

### 3. Stop services

```bash
# Stop all services
make down-all

# Stop and remove volumes (clean everything)
make clean
```

### Advanced: Run services separately

**Backend services only:**
```bash
make build-backend
make up-backend
make down-backend
```

**BFF service only:**
```bash
make build-bff
make up-bff
make down-bff
```

## API Endpoints

The BFF service provides three main endpoints that demonstrate gRPC Federation capabilities:

### GetUserWithOrders
Fetches a user along with their orders by aggregating data from User and Order services.

### GetOrderDetails
Retrieves detailed order information including user details and product information by aggregating data from all three backend services.

### GetUserDashboard
Provides a comprehensive user dashboard with user info, order statistics, recent orders, and product recommendations.

## How gRPC Federation Works

The BFF service uses gRPC Federation options in the Protocol Buffer definitions to:

1. **Define service dependencies**: Specify which backend services are needed
2. **Make parallel/sequential calls**: Automatically optimize service calls
3. **Transform and aggregate data**: Use CEL expressions to process responses
4. **Auto-bind fields**: Map backend responses to BFF response fields

Example from `proto/bff/bff.proto`:

```protobuf
message GetUserWithOrdersResponse {
  option (grpc.federation.message) = {
    def {
      name: "user_res"
      call {
        method: "backend.user.UserService/GetUser"
        request {
          field { field: "user_id" by: "$.user_id" }
        }
      }
    }
    def {
      name: "orders_res"
      call {
        method: "backend.order.OrderService/ListOrdersByUser"
        request {
          field { field: "user_id" by: "$.user_id" }
        }
      }
    }
  };

  backend.user.User user = 1 [(grpc.federation.field).by = "user_res.user"];
  repeated backend.order.Order orders = 2 [(grpc.federation.field).by = "orders_res.orders"];
}
```

## Project Structure

```
grpc-federation-demo/
├── proto/
│   ├── backend/           # Backend service definitions
│   │   ├── user.proto
│   │   ├── product.proto
│   │   └── order.proto
│   └── bff/
│       └── bff.proto      # BFF service with federation
├── backend/               # Backend service implementations
│   ├── user/server.go
│   ├── product/server.go
│   └── order/server.go
└── bff/
    └── server.go          # BFF service implementation
```

## Development

### How to Implement BFF with gRPC Federation

This project demonstrates the typical workflow for implementing a BFF using grpc-federation:

**1. Define your backend services** (`proto/backend/*.proto`)

First, create standard gRPC service definitions for your backend microservices:

```protobuf
// proto/backend/user.proto
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

**2. Write BFF proto with federation annotations** (`proto/bff/bff.proto`)

Add grpc-federation options to define how to aggregate backend services:

```protobuf
service BFFService {
  option (grpc.federation.service) = {
    dependencies: [
      { name: "user" service: "backend.user.UserService" },
      { name: "order" service: "backend.order.OrderService" }
    ]
  };

  rpc GetUserWithOrders(GetUserWithOrdersRequest) returns (GetUserWithOrdersResponse);
}

message GetUserWithOrdersResponse {
  option (grpc.federation.message) = {
    def {
      name: "user_res"
      call { method: "user.UserService/GetUser" ... }
    }
    def {
      name: "orders_res"
      call { method: "order.OrderService/ListOrdersByUser" ... }
    }
  };

  User user = 1 [(grpc.federation.field).by = "user_res.user"];
  repeated Order orders = 2 [(grpc.federation.field).by = "orders_res.orders"];
}
```

**3. Generate Go code**

Run `protoc` with the grpc-federation plugin to auto-generate the BFF implementation:

```bash
make proto-bff
```

This generates:
- `bff.pb.go` - Standard protobuf code
- `bff_grpc.pb.go` - Standard gRPC server/client code
- `bff_grpc_federation.pb.go` - **Federation logic (auto-generated!)**

**4. Implement the BFF server** (`bff/server.go`)

Simply register the auto-generated federation service:

```go
func main() {
    conn := /* create connections to backend services */

    // The federation code handles all the aggregation logic!
    federationServer := bff.NewBFFServiceFederationServer(
        bff.BFFServiceFederationConfig(conn),
    )

    grpc.RegisterService(s, federationServer)
    s.Serve(lis)
}
```

**Key Benefits:**
- ✅ No manual service orchestration code
- ✅ Automatic parallel/sequential call optimization
- ✅ Type-safe field binding
- ✅ Built-in error handling and timeout support

### For Developers

If you want to modify the proto files or service implementations:

```bash
# Initialize project (install required tools)
make init

# Generate proto files after making changes
make proto

# Or generate individually
make proto-backend  # Backend services only
make proto-bff      # BFF service only

# Rebuild and restart services
make build-all
make down-all
make up-all
```

### Available Make Commands

```bash
# Protocol Buffer Generation
make proto              # Generate all proto files
make proto-backend      # Generate backend proto only
make proto-bff          # Generate BFF proto only

# Docker Commands
make build-all          # Build all Docker images
make up-all             # Start all services
make down-all           # Stop all services
make clean              # Stop and remove all containers and volumes

# Testing Commands
make grpcurl-list       # List available services
make grpcurl-user       # Test GetUserWithOrders
make grpcurl-order      # Test GetOrderDetails
make grpcurl-dashboard  # Test GetUserDashboard
make grpcurl-all        # Test all endpoints

# Setup
make init               # Install required tools
```

## Learn More

- [gRPC Federation GitHub](https://github.com/mercari/grpc-federation)
- [Mercari Engineering Blog - gRPC Federation Introduction](https://engineering.mercari.com/blog/entry/20240401-4f426bd460/)
- [Mercari Engineering Blog - Developing BFF using gRPC Federation](https://engineering.mercari.com/blog/entry/20241204-developing-bff-using-grpc-federation/)