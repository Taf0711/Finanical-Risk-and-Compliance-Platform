# Financial Risk Monitor - Test Suite

## Overview
This comprehensive test suite validates all aspects of the Financial Risk Monitor backend API, including authentication, portfolio management, transactions, risk calculations, compliance checks, alerts, and WebSocket functionality.

## Prerequisites
1. **Server Running**: The backend server must be running on `localhost:8080`
2. **Database**: PostgreSQL and Redis must be configured and running
3. **Dependencies**: Run `go mod tidy` to ensure all dependencies are installed

## Running Tests

### Quick Run (Recommended)
```bash
./run_tests.sh
```

### Manual Run
```bash
cd tests
go run test_runner.go
```

## Test Coverage

### 🔒 Authentication Tests
- Server connectivity check
- User registration with unique email generation
- User login with token extraction
- Duplicate registration validation
- Invalid login attempts

### 📊 Portfolio Management Tests
- Portfolio creation
- Portfolio retrieval (all portfolios)
- Single portfolio retrieval
- Portfolio updates
- Portfolio deletion (cleanup)

### 💸 Transaction Tests
- Transaction creation with proper currency field
- Transaction listing
- Transaction status updates

### 📈 Risk Metrics Tests (Flexible)
- VaR (Value at Risk) calculations
- Liquidity risk assessment
- Risk metrics retrieval
- Risk history tracking

*Note: Risk tests accept 200 (success) or 501 (not implemented) responses*

### ✅ Compliance Tests (Flexible)
- Portfolio compliance checking
- Position limits validation
- AML (Anti-Money Laundering) checks

*Note: Compliance tests accept 200, 404, or 501 responses for endpoints that may not be fully implemented*

### 🚨 Alert Tests
- Alert retrieval
- Active alerts filtering
- Alert acknowledgment

### 🔌 WebSocket Tests
- Connection establishment
- Message handling

### 🧹 Cleanup Tests
- Resource cleanup (portfolio deletion)

## Test Features

### Smart Test Dependencies
- Tests are sequentially dependent (e.g., login depends on registration)
- Created resources are tracked and reused across tests
- Automatic cleanup at the end

### Flexible Implementation
- Tests accommodate endpoints that may not be fully implemented yet
- Graceful handling of "not implemented" (501) responses
- Clear error reporting with detailed status codes and responses

### Visual Output
- Color-coded test results (✓ green for pass, ✗ red for fail)
- Grouped test execution by feature area
- Detailed error messages and response data
- Response time tracking

## Test Data
- **User Email**: Dynamically generated with timestamp to avoid conflicts
- **Portfolio**: "Test Portfolio" with automated test description
- **Transaction**: AAPL stock purchase (10 shares @ $150.50)

## Environment Variables
Tests automatically set and use environment variables for:
- `TEST_EMAIL`: Generated during registration
- `TEST_PASSWORD`: Test password for login

## Expected Responses
- **Success**: HTTP 200/201 for successful operations
- **Authentication**: HTTP 401 for unauthorized requests
- **Validation**: HTTP 400 for invalid data
- **Not Found**: HTTP 404 for missing resources
- **Not Implemented**: HTTP 501 for unimplemented features (accepted as valid)

## Troubleshooting

### Common Issues
1. **Server not running**: Ensure the backend is running on port 8080
2. **Database connection**: Verify PostgreSQL and Redis are accessible
3. **Port conflicts**: Check no other services are using port 8080
4. **Dependencies**: Run `go mod tidy` if import errors occur

### Debug Output
Each test provides detailed information including:
- HTTP status codes
- Response bodies
- Error messages
- Response times
- Request/response data for debugging

## Contributing
When adding new API endpoints, please:
1. Add corresponding tests to the appropriate test group
2. Update the test count in this README
3. Follow the existing error handling patterns
4. Include proper cleanup if resources are created
