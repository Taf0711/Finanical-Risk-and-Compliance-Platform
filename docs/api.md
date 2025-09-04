# API Documentation

## Overview

The Financial Risk Monitor API provides endpoints for portfolio management, risk analysis, transaction processing, and real-time monitoring.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Authentication

#### POST /auth/register
Register a new user.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "success": true,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "analyst"
  },
  "token": "jwt_token_here"
}
```

#### POST /auth/login
Authenticate user and get JWT token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123!"
}
```

### Portfolios

#### GET /portfolios
Get all portfolios for the authenticated user.

**Response:**
```json
{
  "portfolios": [
    {
      "id": "uuid",
      "name": "Tech Growth Portfolio",
      "description": "High-growth technology stocks",
      "total_value": "500000.00",
      "currency": "USD",
      "positions": [...],
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /portfolios
Create a new portfolio.

**Request Body:**
```json
{
  "name": "New Portfolio",
  "description": "Portfolio description",
  "currency": "USD"
}
```

#### GET /portfolios/{id}
Get a specific portfolio by ID.

#### PUT /portfolios/{id}
Update a portfolio.

#### DELETE /portfolios/{id}
Delete a portfolio.

### Risk Analysis

#### GET /risk/portfolio/{id}/var
Calculate Value at Risk for a portfolio.

**Response:**
```json
{
  "portfolio_id": "uuid",
  "var_value": "25000.00",
  "var_percentage": 5.0,
  "confidence_level": 95.0,
  "time_horizon": 1,
  "method": "monte_carlo",
  "status": "SAFE",
  "calculated_at": "2024-01-01T00:00:00Z"
}
```

#### GET /risk/portfolio/{id}/liquidity
Calculate liquidity risk for a portfolio.

**Response:**
```json
{
  "portfolio_id": "uuid",
  "liquidity_ratio": 0.75,
  "liquidity_score": "HIGH",
  "days_to_liquidate": 2.0,
  "risk_assessment": "LOW_RISK",
  "calculated_at": "2024-01-01T00:00:00Z"
}
```

### Monte Carlo Simulation

#### POST /monte-carlo/portfolio/{portfolio_id}/simulation
Run Monte Carlo simulation for portfolio risk analysis.

**Request Body:**
```json
{
  "num_simulations": 10000,
  "time_horizon_days": 22,
  "confidence_level": 0.95,
  "market_regime": "NORMAL",
  "correlation_matrix": true
}
```

**Response:**
```json
{
  "success": true,
  "portfolio": {...},
  "simulation": {
    "var_estimates": {
      "VaR_95": 25000.50,
      "VaR_99": 35000.75
    },
    "expected_shortfall": {
      "ES_95": 30000.25,
      "ES_99": 42000.80
    },
    "volatility": 0.22,
    "max_drawdown": 0.15,
    "duration": "2.5s"
  }
}
```

### Transactions

#### GET /transactions
Get all transactions for the user.

#### POST /transactions
Create a new transaction.

**Request Body:**
```json
{
  "portfolio_id": "uuid",
  "transaction_type": "BUY",
  "symbol": "AAPL",
  "quantity": "100.00",
  "price": "150.00",
  "currency": "USD"
}
```

### Alerts

#### GET /alerts
Get all alerts for the user.

#### GET /alerts/active
Get active alerts only.

#### PUT /alerts/{id}/acknowledge
Acknowledge an alert.

#### PUT /alerts/{id}/resolve
Resolve an alert.

## WebSocket

### Connection
```
ws://localhost:8080/ws?user_id=<user_id>
```

### Message Types

#### price_update
Real-time price updates for assets.

```json
{
  "type": "price_update",
  "data": {
    "AAPL": {
      "price": 150.25,
      "change": 1.5,
      "timestamp": 1640995200
    }
  }
}
```

#### risk_update
Real-time risk metric updates.

```json
{
  "type": "risk_update",
  "data": {
    "portfolio_id": "uuid",
    "var": {...},
    "liquidity": {...},
    "timestamp": 1640995200
  }
}
```

#### new_alert
New alert notifications.

```json
{
  "type": "new_alert",
  "data": {
    "alert": {
      "id": "uuid",
      "title": "VaR Limit Exceeded",
      "severity": "HIGH",
      "description": "Portfolio VaR exceeds threshold"
    },
    "timestamp": 1640995200
  }
}
```

## Error Handling

All API endpoints return errors in the following format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {...}
}
```

### Common HTTP Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Validation Error
- `500` - Internal Server Error

## Rate Limiting

API requests are limited to 1000 requests per hour per user.

## Pagination

List endpoints support pagination:

```
GET /portfolios?page=1&limit=10
```

Response includes pagination metadata:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10
  }
}
```
