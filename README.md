# Stock API - Go Backend Service

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A production-ready REST API built with Go for managing stock market data. This service fetches stock information from an external API and stores it in CockroachDB, providing comprehensive CRUD operations with filtering and pagination capabilities.

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```text
stock_backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business entities and interfaces
│   │   ├── stock.go
│   │   └── errors.go
│   ├── usecase/                 # Business logic
│   │   ├── stock_usecase.go
│   │   └── stock_usecase_test.go
│   ├── repository/              # Data access layer
│   │   └── cockroachdb/
│   │       ├── connection.go
│   │       └── stock_repository.go
│   ├── client/                  # External API clients
│   │   └── stock_api_client.go
│   ├── handler/                 # HTTP handlers
│   │   └── stock_handler.go
│   ├── router/                  # Route configuration
│   │   └── router.go
│   ├── middleware/              # HTTP middleware
│   │   └── middleware.go
│   └── config/                  # Configuration management
│       ├── config.go
│       └── config_test.go
├── pkg/                         # Reusable packages
│   └── logger/
│       └── logger.go
├── docs/                        # Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── .gitignore
└── README.md
```

## 🚀 Features

- ✅ **Clean Architecture** - Maintainable and testable code structure
- ✅ **RESTful API** - Standard HTTP methods and status codes
- ✅ **Swagger/OpenAPI** - Interactive API documentation
- ✅ **Stock Recommendations** - Multi-factor scoring algorithm to identify best investment opportunities
- ✅ **Database Integration** - CockroachDB with connection pooling
- ✅ **External API Client** - Fetch stock data from external sources
- ✅ **Smart Deduplication** - Automatically returns only the latest version of each stock (by ticker)
- ✅ **Advanced Filtering** - Filter by ticker, company, brokerage, action, and ratings
- ✅ **Flexible Sorting** - Sort by any field in ascending or descending order
- ✅ **Comprehensive Testing** - Unit tests with mocks
- ✅ **Structured Logging** - JSON logging with Zap
- ✅ **Graceful Shutdown** - Proper cleanup on termination
- ✅ **CORS Support** - Cross-origin resource sharing
- ✅ **Accurate Pagination** - Efficient data retrieval with correct counts
- ✅ **Error Handling** - Consistent error responses
- ✅ **Configuration Management** - Environment-based configuration

## 📋 Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Make** (optional but recommended)
- **CockroachDB** - [Installation Guide](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html)

## 🛠️ Installation

### 1. Clone the repository

```bash
git clone <repository-url>
cd stock_backend
```

### 2. Set up environment variables

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Required: Add your Stock API key
STOCK_API_KEY=your_api_key_here

# Database configuration
DB_HOST=localhost
DB_PORT=26257
DB_NAME=stock_data

# Server configuration
SERVER_PORT=8080
```

### 3. Install dependencies

```bash
go mod download
go mod tidy
```

## 🚦 Running the Application

### Prerequisites

Make sure you have CockroachDB running locally. See [CockroachDB Setup Guide](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html) for installation instructions.

### Start CockroachDB (Single Node)

```bash
# Start CockroachDB in insecure mode (for development)
cockroach start-single-node --insecure --listen-addr=localhost:26257 --http-addr=localhost:8081
```

### Run the Application

```bash
# Run directly
make run

# Or manually
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### Build and Run Binary

```bash
# Build
make build

# Run
./bin/stock-api
```

## 📖 API Documentation

### Swagger UI

Once the application is running, visit:

```text
http://localhost:8080/swagger/index.html
```

### Generate Swagger Documentation

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
make swagger

# Or manually
swag init -g cmd/api/main.go -o docs
```

### Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/stocks` | Get all stocks (with filters, returns latest version per ticker) |
| GET | `/api/v1/stocks/:id` | Get stock by ID |
| GET | `/api/v1/stock/:ticker` | Get all historical versions of a stock by ticker |
| GET | `/api/v1/recommendations` | Get stock investment recommendations based on scoring algorithm |
| POST | `/api/v1/stocks/sync` | Sync stocks from external API |

### Example Requests

#### Sync stocks from external API

```bash
curl -X POST http://localhost:8080/api/v1/stocks/sync
```

#### Get stocks with filters

**Note:** The API automatically returns only the **latest version** of each stock (by ticker). When stocks are synchronized from the external API multiple times, they may have different timestamps. The API intelligently filters these duplicates and returns only the most recent entry for each ticker, ensuring accurate pagination counts.

```bash
# Get all stocks (paginated)
curl http://localhost:8080/api/v1/stocks?limit=10&offset=0

# Filter by ticker
curl http://localhost:8080/api/v1/stocks?ticker=AAPL

# Filter by company (partial match)
curl http://localhost:8080/api/v1/stocks?company=Apple

# Filter by brokerage
curl http://localhost:8080/api/v1/stocks?brokerage=Morgan

# Filter by action
curl http://localhost:8080/api/v1/stocks?action=upgrade

# Filter by rating_from (original rating)
curl http://localhost:8080/api/v1/stocks?rating_from=Neutral

# Filter by rating_to (target rating)
curl http://localhost:8080/api/v1/stocks?rating_to=Overweight

# Filter by both ratings
curl "http://localhost:8080/api/v1/stocks?rating_from=Neutral&rating_to=Overweight"

# Sort by ticker (ascending)
curl "http://localhost:8080/api/v1/stocks?sortBy=ticker&sortOrder=asc"

# Sort by company name
curl "http://localhost:8080/api/v1/stocks?sortBy=company&sortOrder=asc"

# Sort by time (default: descending - newest first)
curl "http://localhost:8080/api/v1/stocks?sortBy=time&sortOrder=desc"

# Multiple filters with sorting and pagination
curl "http://localhost:8080/api/v1/stocks?ticker=AAPL&sortBy=time&sortOrder=desc&limit=5&offset=0"

# Complex query: filter by company and rating, sort by time
curl "http://localhost:8080/api/v1/stocks?company=Apple&rating_to=Overweight&sortBy=time&sortOrder=desc&limit=10"
```

**Available Query Parameters:**
- `ticker` - Filter by exact ticker symbol (e.g., AAPL, GOOGL)
- `company` - Filter by company name (partial match, case-insensitive)
- `brokerage` - Filter by brokerage name (partial match, case-insensitive)
- `action` - Filter by action type (e.g., upgrade, downgrade, initiated)
- `rating_from` - Filter by original rating
- `rating_to` - Filter by target rating
- `sortBy` - Sort field: `ticker`, `company`, `time`, `rating_to`, `action`, `brokerage`, `target_to` (default: `time`)
- `sortOrder` - Sort direction: `asc` or `desc` (default: `desc`)
- `limit` - Number of items per page (default: 50)
- `offset` - Number of items to skip for pagination (default: 0)

#### Get stock by ID

```bash
curl http://localhost:8080/api/v1/stocks/1
```

#### Get all historical versions of a stock by ticker

This endpoint returns **all historical records** for a specific ticker, ordered by time (newest first). Unlike the `/stocks` endpoint which returns only the latest version, this endpoint gives you the complete history.

```bash
# Get all historical versions of Apple stock
curl http://localhost:8080/api/v1/stock/AAPL

# Get all historical versions of AMD stock
curl http://localhost:8080/api/v1/stock/AMD

# Get all historical versions of Tesla stock
curl http://localhost:8080/api/v1/stock/TSLA
```

**Example Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "1111776686872650500",
      "ticker": "AAPL",
      "target_from": "$200.00",
      "target_to": "$220.00",
      "company": "Apple Inc.",
      "action": "upgrade",
      "brokerage": "Goldman Sachs",
      "rating_from": "Neutral",
      "rating_to": "Buy",
      "time": "2025-10-04T10:00:00Z",
      "created_at": "2025-10-04T10:05:00Z",
      "updated_at": "2025-10-04T10:05:00Z"
    },
    {
      "id": "1111776413695180800",
      "ticker": "AAPL",
      "target_from": "$190.00",
      "target_to": "$200.00",
      "company": "Apple Inc.",
      "action": "maintained",
      "brokerage": "Morgan Stanley",
      "rating_from": "Buy",
      "rating_to": "Buy",
      "time": "2025-10-01T14:30:00Z",
      "created_at": "2025-10-01T14:35:00Z",
      "updated_at": "2025-10-01T14:35:00Z"
    }
  ]
}
```

#### Get stock recommendations

This endpoint analyzes all stock data and returns the best investment recommendations based on a sophisticated scoring algorithm.

```bash
# Get top 10 recommendations (default)
curl http://localhost:8080/api/v1/recommendations

# Get top 5 recommendations
curl http://localhost:8080/api/v1/recommendations?limit=5

# Get top 20 recommendations
curl http://localhost:8080/api/v1/recommendations?limit=20
```

**Scoring Algorithm:**

The recommendation engine evaluates each stock based on multiple weighted factors:

| Factor | Weight | Description |
|--------|--------|-------------|
| **Action Type** | 30% | Upgrades score highest, downgrades score lowest |
| **Rating Improvement** | 25% | Improvement from rating_from to rating_to (e.g., Neutral → Buy) |
| **Target Price Increase** | 20% | Percentage increase from target_from to target_to |
| **Recency** | 15% | More recent ratings score higher |
| **Brokerage Reputation** | 10% | Top-tier brokerages (Goldman Sachs, Morgan Stanley) score higher |

**Action Scores:**
- Upgrade: 10.0
- Initiated Coverage: 8.0
- Target Raised: 7.0
- Reiterated/Maintained: 6.0
- Target Lowered: 3.0
- Downgrade: 2.0

**Rating Hierarchy (1-5 scale):**
- **Strong Buy**: 5
- **Buy / Speculative Buy / Overweight / Outperform / Market Outperform / Sector Outperform / Positive**: 4
- **Hold / Neutral / In-Line / Market Perform / Sector Perform / Equal Weight**: 3
- **Underweight / Underperform / Reduce**: 2
- **Sell**: 1

**Rating Improvement Bonus:**
- Upgrade (e.g., Neutral → Buy): +4 points
- Downgrade (e.g., Buy → Hold): -4 points
- Maintained rating: No bonus/penalty

**Example Response:**
```json
{
  "success": true,
  "message": "Top 10 stock recommendations based on recent ratings, actions, and target prices",
  "data": [
    {
      "stock": {
        "id": "1111776686872650500",
        "ticker": "NVDA",
        "target_from": "$450.00",
        "target_to": "$550.00",
        "company": "NVIDIA Corporation",
        "action": "upgrade",
        "brokerage": "Goldman Sachs",
        "rating_from": "Neutral",
        "rating_to": "Buy",
        "time": "2025-10-04T09:00:00Z"
      },
      "score": 8.95,
      "reason": "Recent upgrade; Rating improved to Buy; 22.2% price target increase; Rated by Goldman Sachs",
      "target_increase_percent": 22.2
    },
    {
      "stock": {
        "id": "1111776413695180801",
        "ticker": "TSLA",
        "target_from": "$240.00",
        "target_to": "$300.00",
        "company": "Tesla Inc.",
        "action": "initiated coverage",
        "brokerage": "Morgan Stanley",
        "rating_from": "",
        "rating_to": "Overweight",
        "time": "2025-10-03T14:30:00Z"
      },
      "score": 8.52,
      "reason": "Recent initiated coverage; Rating improved to Overweight; 25.0% price target increase; Rated by Morgan Stanley",
      "target_increase_percent": 25.0
    }
  ]
}
```

**Use Cases:**
- Quick investment ideas based on recent analyst ratings
- Identify stocks with strong upward momentum
- Filter high-conviction recommendations from top brokerages
- Compare multiple opportunities at once

## 🧪 Testing

### Run all tests

```bash
make test
```

### Run tests in watch mode

```bash
make test-watch
```

### Run tests with coverage

```bash
make test-coverage
```

### View coverage report in browser

```bash
make test-coverage-html
```

### Run specific test

```bash
go test -v ./internal/usecase/...
```

## 🏗️ Building

### Build for current platform

```bash
make build
```

### Cross-platform builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/stock-api-linux cmd/api/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/stock-api.exe cmd/api/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/stock-api-macos cmd/api/main.go
```

## 🔧 Development

### Project Structure Explanation

- **`cmd/`** - Application entry points
- **`internal/`** - Private application code
  - **`domain/`** - Business entities and repository interfaces
  - **`usecase/`** - Business logic implementation
  - **`repository/`** - Data persistence implementations
  - **`client/`** - External service clients
  - **`handler/`** - HTTP request handlers
  - **`router/`** - Route definitions
  - **`middleware/`** - HTTP middleware
  - **`config/`** - Configuration management
- **`pkg/`** - Public reusable packages
- **`docs/`** - API documentation

### Code Quality

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run all checks
make check
```

### Database Migrations

The application automatically creates the database schema on startup. The schema includes:

- **`stocks`** table with indexes on ticker, company, time, and brokerage
- Unique constraint on (ticker, company, time) to prevent duplicates

## 📊 Monitoring and Logging

### Logs

The application uses structured JSON logging with different levels:

- **DEBUG** - Detailed information for debugging
- **INFO** - General informational messages
- **WARN** - Warning messages
- **ERROR** - Error messages
- **FATAL** - Fatal errors that cause the application to exit

### Health Check

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "success": true,
  "message": "Service is healthy"
}
```

## 🔒 Security Considerations

- ✅ API keys stored in environment variables
- ✅ SQL injection prevention using parameterized queries
- ✅ Connection pooling to prevent resource exhaustion
- ✅ Request timeouts to prevent hanging connections
- ✅ CORS configuration for cross-origin requests