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
- ✅ **Database Integration** - CockroachDB with connection pooling
- ✅ **External API Client** - Fetch stock data from external sources
- ✅ **Comprehensive Testing** - Unit tests with mocks
- ✅ **Structured Logging** - JSON logging with Zap
- ✅ **Graceful Shutdown** - Proper cleanup on termination
- ✅ **CORS Support** - Cross-origin resource sharing
- ✅ **Pagination** - Efficient data retrieval
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
| GET | `/api/v1/stocks` | Get all stocks (with filters) |
| GET | `/api/v1/stocks/:id` | Get stock by ID |
| POST | `/api/v1/stocks/sync` | Sync stocks from external API |

### Example Requests

#### Sync stocks from external API

```bash
curl -X POST http://localhost:8080/api/v1/stocks/sync
```

#### Get stocks with filters

```bash
# Get all stocks (paginated)
curl http://localhost:8080/api/v1/stocks?limit=10&offset=0

# Filter by ticker
curl http://localhost:8080/api/v1/stocks?ticker=AAPL

# Filter by company (partial match)
curl http://localhost:8080/api/v1/stocks?company=Apple

# Multiple filters
curl "http://localhost:8080/api/v1/stocks?ticker=AAPL&limit=5"
```

#### Get stock by ID

```bash
curl http://localhost:8080/api/v1/stocks/1
```

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

### CockroachDB Admin UI

Access the CockroachDB admin UI at:

```text
http://localhost:8081
```

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

## 🚀 Deployment

### Production Checklist

- [ ] Set `ENV=production` in environment variables
- [ ] Configure SSL/TLS for database connection
- [ ] Set appropriate connection pool sizes
- [ ] Configure log level to `info` or `warn`
- [ ] Set up monitoring and alerting
- [ ] Configure backup strategy for database
- [ ] Use secrets management for API keys
- [ ] Set up reverse proxy (nginx/Caddy) for SSL termination

### Environment Variables (Production)

```env
ENV=production
SERVER_PORT=8080
LOG_LEVEL=info
LOG_FORMAT=json
DB_SSLMODE=require
DB_MAX_CONNS=50
STOCK_API_TIMEOUT=30s
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Write unit tests for new features
- Update documentation as needed
- Use meaningful commit messages
- Keep functions small and focused

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- Your Name - Initial work

## 🙏 Acknowledgments

- Go community for excellent tooling
- CockroachDB team for the distributed SQL database
- Gin framework for the HTTP router
- Zap for structured logging
- Testify for testing utilities

## 📞 Support

For issues and questions:

- Create an issue in the repository
- Contact the development team

## 🗺️ Roadmap

- [ ] Add authentication and authorization
- [ ] Implement rate limiting
- [ ] Add caching layer (Redis)
- [ ] Implement WebSocket for real-time updates
- [ ] Add GraphQL support
- [ ] Implement background job processing
- [ ] Add metrics and tracing (Prometheus, Jaeger)
- [ ] Implement CI/CD pipeline

---

Built with ❤️ using Go
