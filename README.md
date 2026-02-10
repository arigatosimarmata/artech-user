# Artech User Microservice

A production-ready User Microservice built with Go, implementing Clean Architecture principles with comprehensive authentication, authorization, and audit trail features.

## 🚀 Features

### Authentication & Authorization
- ✅ User Registration with email validation
- ✅ User Login with JWT tokenization
- ✅ Token Refresh mechanism
- ✅ Logout with token revocation
- ✅ Forgot Password with email notification
- ✅ Password Reset with token validation
- ✅ OAuth structure preparation (Google, Facebook, GitHub)

### Security
- ✅ Password hashing with bcrypt
- ✅ JWT-based authentication
- ✅ Token expiration and refresh
- ✅ Token revocation on logout
- ✅ Secure password reset flow

### Audit Trail
- ✅ Track all user activities
- ✅ Record IP address and User-Agent
- ✅ Store old and new values for changes
- ✅ Query audit logs by user

### Architecture
- ✅ Clean Architecture with clear layer separation
- ✅ Dependency injection using composition
- ✅ Interface-based design for testability
- ✅ Structured logging with Zap
- ✅ Error handling with custom error types
- ✅ Request validation at DTO layer

## 📋 Technology Stack

- **Language**: Go 1.21
- **HTTP Framework**: Fiber v2
- **Database**: MySQL 8.0
- **Database Library**: sqlx (raw SQL queries)
- **CLI Framework**: Cobra
- **JWT**: golang-jwt/jwt
- **Logging**: Uber Zap
- **Password Hashing**: bcrypt
- **Validation**: go-playground/validator

## 🏗️ Architecture

This project follows Clean Architecture principles with the following layers:

```
┌─────────────────────────────────────────────────────────────┐
│                      Presentation Layer                      │
│                  (Handler, Middleware)                       │
├─────────────────────────────────────────────────────────────┤
│                      Application Layer                       │
│                       (Use Cases)                            │
├─────────────────────────────────────────────────────────────┤
│                       Domain Layer                           │
│                (Entities, Errors, Interfaces)                │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│             (Repository, Config, External Services)          │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
artech-user/
├── cmd/
│   ├── migrate/          # Migration CLI entrypoint
│   └── http/             # HTTP server entrypoint
├── internal/
│   ├── config/           # Configuration management
│   ├── middleware/       # HTTP middleware
│   └── migration/        # Migration logic
├── domain/               # Domain entities and errors
├── dto/                  # Data Transfer Objects
│   ├── request/          # Request DTOs
│   └── response/         # Response DTOs
├── repository/           # Data access layer
├── usecase/              # Business logic layer
├── handler/              # HTTP handlers
├── pkg/                  # Shared utilities
│   ├── logger/           # Logging utility
│   ├── token/            # JWT utility
│   ├── password/         # Password hashing
│   ├── email/            # Email sender
│   └── audit/            # Audit trail helper
├── migration/
│   └── sql/              # SQL migration files
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

## 🗄️ Database Schema

### Tables

1. **users** - User accounts
2. **user_tokens** - Access and refresh tokens
3. **oauth_accounts** - OAuth provider accounts
4. **password_reset_tokens** - Password reset tokens
5. **audit_trails** - Activity audit logs

All tables have proper indexing for optimal query performance.

## 🚦 Getting Started

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0 or higher
- Docker & Docker Compose (optional)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/arigatosimarmata/artech-user.git
   cd artech-user
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run with Docker Compose (Recommended)**
   ```bash
   docker-compose up -d
   ```
   This will start:
   - MySQL database on port 3306
   - PHPMyAdmin on port 8081
   - Application server on port 8080

5. **Or run manually**
   
   Start MySQL:
   ```bash
   # Make sure MySQL is running
   ```
   
   Run migrations:
   ```bash
   go run cmd/migrate/main.go up
   ```
   
   Start server:
   ```bash
   go run cmd/http/main.go start
   ```

## 🔧 CLI Commands

### Migration CLI

```bash
# Apply all pending migrations
go run cmd/migrate/main.go up

# Show migration status
go run cmd/migrate/main.go status

# Rollback last migration
go run cmd/migrate/main.go down

# Specify migration directory
go run cmd/migrate/main.go up --dir ./migration/sql
```

### HTTP Server CLI

```bash
# Start HTTP server
go run cmd/http/main.go start
```

## 📡 API Endpoints

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login user |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password |

### Authentication (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/logout` | Logout user |

### User (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/user/profile` | Get user profile |

### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

## 📝 API Examples

### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "full_name": "John Doe",
    "phone": "1234567890"
  }'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get Profile (Protected)

```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Logout (Protected)

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Forgot Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com"
  }'
```

### Reset Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "RESET_TOKEN",
    "new_password": "newpassword123"
  }'
```

## 🔐 Environment Variables

See `.env.example` for all available configuration options:

- **Server Configuration**: Host, port, timeouts
- **Database Configuration**: Connection details, pool settings
- **JWT Configuration**: Secret, token durations
- **Logger Configuration**: Level, file path, rotation settings
- **Email Configuration**: SMTP settings

## 🏗️ Building for Production

### Build Binaries

```bash
# Build migration CLI
go build -o bin/migrate cmd/migrate/main.go

# Build HTTP server
go build -o bin/server cmd/http/main.go
```

### Build Docker Image

```bash
docker build -t artech-user:latest .
```

## 🧪 Development

### Code Structure Guidelines

1. **Domain Layer**: Contains entities and business rules
2. **DTO Layer**: Request/response data structures with validation
3. **Repository Layer**: Database operations only
4. **Use Case Layer**: Business logic implementation
5. **Handler Layer**: HTTP request/response handling
6. **Middleware Layer**: Cross-cutting concerns

### Error Handling

- Use domain errors for business logic errors
- Return proper HTTP status codes
- Log errors with context
- Never expose internal errors to clients

### Logging

- Use structured logging with Zap
- Log at appropriate levels (Info, Warn, Error, Debug)
- Include contextual information
- Rotate logs automatically

### Audit Trail

- All user activities are automatically logged
- Tracks IP address and User-Agent
- Records old and new values for changes
- Queryable by user ID

## 🔒 Security Features

- Password hashing with bcrypt (cost 10)
- JWT-based authentication
- Token expiration and refresh mechanism
- Secure password reset flow
- Token revocation on logout
- Request validation
- SQL injection prevention (parameterized queries)
- CORS configuration
- Rate limiting ready structure

## 📦 Dependencies

See `go.mod` for complete list. Key dependencies:

- `github.com/gofiber/fiber/v2` - HTTP framework
- `github.com/jmoiron/sqlx` - Database library
- `github.com/spf13/cobra` - CLI framework
- `github.com/golang-jwt/jwt/v5` - JWT library
- `go.uber.org/zap` - Logging library
- `golang.org/x/crypto` - Cryptography
- `github.com/go-playground/validator/v10` - Validation

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👨‍💻 Author

**arigatosimarmata**

## 🙏 Acknowledgments

- Clean Architecture by Robert C. Martin
- Go best practices
- Fiber framework community
