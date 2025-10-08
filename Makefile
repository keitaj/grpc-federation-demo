.PHONY: proto proto-backend proto-bff init \
	build-backend build-bff build-all \
	up-backend up-bff up-all \
	down-backend down-bff down-all clean \
	grpcurl-list grpcurl-user grpcurl-order grpcurl-dashboard grpcurl-all

# ========================================
# Protocol Buffer Generation
# ========================================

# Generate all protocol buffers
proto: proto-backend proto-bff

# Generate backend service protocol buffers
proto-backend:
	protoc -I. \
		-I$(shell go list -m -f '{{.Dir}}' github.com/mercari/grpc-federation)/proto_deps \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/backend/*.proto

# Generate BFF service with federation
proto-bff:
	protoc -I. \
		-I$(shell go list -m -f '{{.Dir}}' github.com/mercari/grpc-federation)/proto \
		-I$(shell go list -m -f '{{.Dir}}' github.com/mercari/grpc-federation)/proto_deps \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-federation_out=. --grpc-federation_opt=paths=source_relative \
		proto/bff/*.proto

# ========================================
# Setup Commands
# ========================================

# Initialize project (install tools and setup modules)
init:
	@echo "Installing required Go tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/mercari/grpc-federation/cmd/protoc-gen-grpc-federation@latest
	@echo ""
	@echo "Initializing Go modules..."
	go mod tidy
	go mod download
	@echo ""
	@echo "Setup complete!"

# ========================================
# Docker Compose Commands
# ========================================

# Docker Compose commands for backend services
build-backend:
	docker-compose -f docker-compose.backend.yml build

up-backend:
	docker-compose -f docker-compose.backend.yml up -d

down-backend:
	docker-compose -f docker-compose.backend.yml down

# Docker Compose commands for BFF service
build-bff:
	docker-compose -f docker-compose.bff.yml build

up-bff:
	docker-compose -f docker-compose.bff.yml up -d

down-bff:
	docker-compose -f docker-compose.bff.yml down

# Docker Compose commands for all services
build-all:
	docker-compose build

up-all:
	docker-compose up -d

down-all:
	docker-compose down

# Stop and remove all containers, networks, and volumes
clean:
	docker-compose down -v
	docker-compose -f docker-compose.backend.yml down -v
	docker-compose -f docker-compose.bff.yml down -v

# ========================================
# grpcurl Test Commands
# ========================================

# grpcurl commands to test BFF service
grpcurl-list:
	@echo "=== Available Services ==="
	@grpcurl -plaintext localhost:50054 list
	@echo ""
	@echo "=== BFF Service Methods ==="
	@grpcurl -plaintext localhost:50054 list bff.BFFService

grpcurl-user:
	@echo "=== GetUserWithOrders - Success Case ==="
	@echo "Testing with user-001 (normal user with orders)"
	@grpcurl -plaintext -d '{"user_id": "user-001"}' localhost:50054 bff.BFFService/GetUserWithOrders
	@echo ""
	@echo "=== GetUserWithOrders - FailedPrecondition: USER_NOT_FOUND ==="
	@echo "Testing with user-notfound"
	@grpcurl -plaintext -d '{"user_id": "user-notfound"}' localhost:50054 bff.BFFService/GetUserWithOrders || true
	@echo ""
	@echo "=== GetUserWithOrders - FailedPrecondition: USER_ACCOUNT_SUSPENDED ==="
	@echo "Testing with user-999 (suspended account)"
	@grpcurl -plaintext -d '{"user_id": "user-999"}' localhost:50054 bff.BFFService/GetUserWithOrders || true

grpcurl-order:
	@echo "=== GetOrderDetails - Success Case ==="
	@echo "Testing with order-001 (normal order)"
	@grpcurl -plaintext -d '{"order_id": "order-001"}' localhost:50054 bff.BFFService/GetOrderDetails
	@echo ""
	@echo "=== GetOrderDetails - FailedPrecondition: ORDER_NOT_FOUND ==="
	@echo "Testing with order-notfound"
	@grpcurl -plaintext -d '{"order_id": "order-notfound"}' localhost:50054 bff.BFFService/GetOrderDetails || true
	@echo ""
	@echo "=== GetOrderDetails - FailedPrecondition: ORDER_CANCELLED ==="
	@echo "Testing with order-999 (cancelled order)"
	@grpcurl -plaintext -d '{"order_id": "order-999"}' localhost:50054 bff.BFFService/GetOrderDetails || true
	@echo ""
	@echo "=== GetOrderDetails - FailedPrecondition: PRODUCT_NOT_FOUND ==="
	@echo "Testing with order-998 (order with non-existent product)"
	@grpcurl -plaintext -d '{"order_id": "order-998"}' localhost:50054 bff.BFFService/GetOrderDetails || true
	@echo ""
	@echo "=== GetOrderDetails - FailedPrecondition: PRODUCT_OUT_OF_STOCK ==="
	@echo "Testing with order-997 (order with out-of-stock product)"
	@grpcurl -plaintext -d '{"order_id": "order-997"}' localhost:50054 bff.BFFService/GetOrderDetails || true

grpcurl-dashboard:
	@echo "=== GetUserDashboard - Success Case ==="
	@echo "Testing with user-002 (normal user)"
	@grpcurl -plaintext -d '{"user_id": "user-002"}' localhost:50054 bff.BFFService/GetUserDashboard
	@echo ""
	@echo "=== GetUserDashboard - FailedPrecondition: USER_NOT_FOUND ==="
	@echo "Testing with user-notfound"
	@grpcurl -plaintext -d '{"user_id": "user-notfound"}' localhost:50054 bff.BFFService/GetUserDashboard || true
	@echo ""
	@echo "=== GetUserDashboard - FailedPrecondition: USER_ACCOUNT_SUSPENDED ==="
	@echo "Testing with user-999 (suspended account)"
	@grpcurl -plaintext -d '{"user_id": "user-999"}' localhost:50054 bff.BFFService/GetUserDashboard || true

grpcurl-all:
	@echo "=========================================="
	@echo "Testing all BFF APIs with grpcurl"
	@echo "=========================================="
	@echo ""
	@$(MAKE) -s grpcurl-user
	@echo ""
	@echo "=========================================="
	@echo ""
	@$(MAKE) -s grpcurl-order
	@echo ""
	@echo "=========================================="
	@echo ""
	@$(MAKE) -s grpcurl-dashboard
	@echo ""
	@echo "=========================================="
	@echo "All tests completed!"
	@echo "=========================================="