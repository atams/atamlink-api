# AtamLink Catalog Service

Service untuk mengelola katalog digital bisnis dengan fitur multi-tenant.

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Make (optional tapi recommended)

### Setup Development

1. Clone repository

```bash
git clone https://github.com/atam/atamlink-api.git
cd atamlink-api
```

2. Copy environment file

```bash
cp .env.example .env
```

3. Install dependencies

```bash
go mod download
```

4. Setup database

```bash
# Start PostgreSQL container (optional)
make db-up

# Run migrations
make migrate-up
```

5. Run service

```bash
# Development dengan hot reload
make dev

# atau langsung
make run
```

Service akan berjalan di `http://localhost:8080`

## 📁 Project Structure

```
.
├── cmd/
│   └── catalogd/          # Application entrypoint
├── internal/
│   ├── config/            # Configuration
│   ├── constant/          # Constants (errors, roles, status)
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # HTTP middlewares
│   ├── service/           # Shared services
│   ├── mod_business/      # Business module
│   ├── mod_catalog/       # Catalog module
│   ├── mod_master/        # Master data module
│   └── mod_user/          # User module
├── pkg/
│   ├── database/          # Database utilities
│   ├── errors/            # Error handling
│   ├── logger/            # Logging
│   └── utils/             # Common utilities
└── internal/database/
    └── migrations/        # Database migrations
```

### Module Structure

Setiap module memiliki struktur:

```
mod_xxx/
├── dto/          # Data Transfer Objects
├── entity/       # Domain entities
├── repository/   # Data access layer
└── usecase/      # Business logic
```

## 🔧 Configuration

### Environment Variables

| Variable      | Description                   | Default             |
| ------------- | ----------------------------- | ------------------- |
| `SERVER_PORT` | Server port                   | `8080`              |
| `SERVER_MODE` | Gin mode (debug/release)      | `debug`             |
| `DB_HOST`     | PostgreSQL host               | `localhost`         |
| `DB_PORT`     | PostgreSQL port               | `5432`              |
| `DB_USER`     | PostgreSQL user               | `atamlink_user`     |
| `DB_PASSWORD` | PostgreSQL password           | `atamlink_password` |
| `DB_NAME`     | PostgreSQL database           | `atamlink_db`       |
| `AUTH_BYPASS` | Bypass auth untuk development | `false`             |

Lihat `.env.example` untuk daftar lengkap.

## 🛠 Development

### Available Commands

```bash
# Development
make dev          # Run dengan hot reload
make run          # Run normal
make build        # Build binary
make test         # Run tests

# Database
make db-up        # Start PostgreSQL container
make db-down      # Stop PostgreSQL container
make db-shell     # Access PostgreSQL shell
make migrate-up   # Run migrations
make migrate-down # Rollback migrations

# Others
make fmt          # Format code
make lint         # Run linter
make help         # Show all commands
```

### Testing with Auth Bypass

Untuk development, aktifkan auth bypass:

```env
AUTH_BYPASS=true
AUTH_BYPASS_USER_ID=550e8400-e29b-41d4-a716-446655440000
AUTH_BYPASS_PROFILE_ID=1
```

## 📚 API Documentation

### Health Check

```bash
# Basic health check
GET /health

# Database health check
GET /health/db
```

### Business Management

```bash
# List businesses
GET    /api/v1/businesses

# Create business
POST   /api/v1/businesses

# Get business detail
GET    /api/v1/businesses/:id

# Update business
PUT    /api/v1/businesses/:id

# Delete business
DELETE /api/v1/businesses/:id
```

### Catalog Management

```bash
# List catalogs
GET    /api/v1/catalogs

# Create catalog
POST   /api/v1/catalogs

# Get catalog detail
GET    /api/v1/catalogs/:id

# Update catalog
PUT    /api/v1/catalogs/:id

# Delete catalog
DELETE /api/v1/catalogs/:id
```

### Master Data

```bash
# List plans
GET /api/v1/masters/plans

# List themes
GET /api/v1/masters/themes
```

## 🏗 Architecture

### Clean Architecture

Project ini mengikuti prinsip Clean Architecture:

1. **Entity Layer**: Domain models dan business rules
2. **Use Case Layer**: Application business logic
3. **Repository Layer**: Data access abstraction
4. **Handler Layer**: HTTP request/response handling

### Database Schema

- **Users & Profiles**: User authentication dan profile management
- **Businesses**: Multi-tenant business management
- **Catalogs**: Digital catalog dengan sections dinamis
- **Master Data**: Plans dan themes

### Key Features

- ✅ Multi-tenant architecture
- ✅ Role-based access control
- ✅ Dynamic catalog sections
- ✅ File upload management
- ✅ Subscription management
- ✅ Slug generation
- ✅ Raw SQL (no ORM)

## 🚦 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific test
go test ./internal/mod_business/...
```

## 🚀 Deployment

### Build Production

```bash
# Build binary
make build

# Build docker image
make docker-build
```

### Environment Setup

1. Setup PostgreSQL database
2. Run migrations
3. Configure environment variables
4. Run the service

## 📝 Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Create Pull Request

### Code Style

- Follow Go conventions
- Use `make fmt` before commit
- Add tests for new features
- Update documentation

## 📄 License

[MIT License](LICENSE)

## 👥 Team

- Backend: Golang + Gin Framework
- Database: PostgreSQL
- Architecture: Clean Architecture

---

**Note**: Ini adalah MVP version. Fitur seperti caching (Redis), message queue, dan monitoring akan ditambahkan di fase berikutnya.
