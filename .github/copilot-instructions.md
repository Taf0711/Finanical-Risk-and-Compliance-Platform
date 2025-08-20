# GitHub Copilot Instructions

## Architecture Overview

This is a financial risk monitoring platform built with Go/Fiber backend. The system follows a clean architecture pattern with clear separation of concerns:

### Service Architecture
- **Handlers** (`internal/handlers/`): HTTP request handlers using Fiber framework
- **Services** (`internal/services/`): Business logic layer with database operations
- **Models** (`internal/models/`): GORM data models with UUID primary keys and decimal fields for financial data
- **Middleware** (`internal/middleware/`): JWT authentication, role-based access control
- **WebSocket** (`internal/websocket/`): Real-time communication using Gorilla WebSocket

### Key Technologies
- **Framework**: Fiber v2 for HTTP routing
- **Database**: PostgreSQL with GORM ORM, Redis for caching
- **Auth**: JWT tokens with role-based middleware (`analyst`, `admin`, `trader`)
- **Decimals**: `shopspring/decimal` for precise financial calculations
- **UUIDs**: Google UUID for all primary keys
- **WebSocket**: Gorilla WebSocket for real-time updates

## Development Workflows

### Server Management
```bash
# Build and run server
cd backend && go build -o financial-risk-monitor ./cmd/api
./financial-risk-monitor

# Kill existing server
lsof -ti:8080 | xargs kill -9

# Run tests
cd backend && go run ./tests/test_runner.go
```

### Database Patterns
- All models use `uuid.UUID` primary keys with `BeforeCreate` hooks
- Financial amounts use `decimal.Decimal` with precise scaling
- Soft deletes enabled on User model with `gorm.DeletedAt`
- Foreign key relationships with proper GORM tags

## Project-Specific Conventions

### Handler Pattern
Handlers inject services in constructors and extract user context from JWT middleware:
```go
func (h *PortfolioHandler) CreatePortfolio(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string) // From JWT middleware
    // Parse request, call service, return JSON response
}
```

### Service Layer Pattern
Services use dependency injection with database singleton:
```go
type PortfolioService struct {
    db *gorm.DB
}

func NewPortfolioService() *PortfolioService {
    return &PortfolioService{db: database.GetDB()}
}
```

### Authentication Flow
1. JWT middleware extracts `user_id`, `user_email`, `user_role` from tokens
2. Role middleware (`AdminMiddleware()`, `RoleMiddleware("analyst")`) for authorization
3. All protected routes use `middleware.JWTMiddleware(authService)`

### Error Handling Convention
Consistent JSON error responses with proper HTTP status codes:
```go
return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
    "error": "descriptive error message",
})
```

## Integration Points

### WebSocket Real-time Updates
- Hub pattern manages client connections
- Messages broadcast to all connected clients
- WebSocket endpoint: `/ws` with proper upgrade handling
- Client registration with unique client IDs

### Configuration Management
Environment-based config loading from `.env` files with structured config types:
- `AppConfig`, `DatabaseConfig`, `RedisConfig`, `JWTConfig`, etc.
- Config loaded once at startup in `main.go`

### Testing Infrastructure
Comprehensive test suite (`tests/test_runner.go`) validates:
- Authentication flows with token extraction
- CRUD operations across all endpoints
- WebSocket connectivity and messaging
- Compliance and risk calculation endpoints

## Critical File Locations
- **Main entry**: `cmd/api/main.go` - server setup, routing, middleware chain
- **Models**: `internal/models/*.go` - GORM models with relationships
- **Database**: `internal/database/postgres.go` - connection and migration
- **Auth**: `internal/services/auth.go` - JWT generation/validation
- **Tests**: `tests/test_runner.go` - comprehensive API validation
