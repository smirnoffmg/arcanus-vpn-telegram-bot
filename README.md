# Arcanus VPN Telegram Bot

A Telegram bot for VPN service management with event-driven architecture and user quota tracking.

## Purpose

- User registration and trial activation (50MB free quota)
- VPN account management through Telegram interface
- Real-time usage tracking and notifications
- Event sourcing with Kafka for audit trail and analytics

## Architecture

### Core Components

```mermaid
graph TB
    subgraph "External"
        TG[Telegram Bot API]
    end
    
    subgraph "Application Layer"
        BOT[Telegram Bot]
        HANDLERS[Handlers]
        SERVICES[Services]
    end
    
    subgraph "Infrastructure Layer"
        EVENTS[Event System]
        REPO[Repository]
    end
    
    subgraph "External Services"
        KAFKA[Kafka]
        DB[(PostgreSQL)]
    end
    
    TG --> BOT
    BOT --> HANDLERS
    HANDLERS --> SERVICES
    SERVICES --> REPO
    REPO --> DB
    SERVICES --> EVENTS
    EVENTS --> KAFKA
    
    style TG fill:#0088cc
    style BOT fill:#f0f0f0
    style HANDLERS fill:#e1f5fe
    style SERVICES fill:#e8f5e8
    style EVENTS fill:#fff3e0
    style REPO fill:#f3e5f5
    style KAFKA fill:#ff9800
    style DB fill:#4caf50
```

### Event System

All user state changes are published to Kafka with user-based partitioning:

- `user.registered` - New user registration
- `user.trial_activated` - Trial activation
- `user.quota_updated` - Quota usage changes
- `bot.message_received` - User interactions
- `system.*` - Application lifecycle events

### Technology Stack

- **Go 1.25** - Backend service
- **PostgreSQL** - Primary data store  
- **Kafka** - Event streaming and audit log
- **GORM** - Database ORM
- **Uber FX** - Dependency injection
- **Logrus** - Structured logging

## Configuration

### Environment Variables

| Variable             | Description                                  | Required |
| -------------------- | -------------------------------------------- | -------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather                    | Yes      |
| `DATABASE_URL`       | PostgreSQL connection string                 | Yes      |
| `KAFKA_BROKERS`      | Kafka broker addresses                       | No*      |
| `KAFKA_TOPIC`        | Event topic name                             | No*      |
| `KAFKA_ENABLED`      | Enable/disable event publishing              | No       |
| `LOG_LEVEL`          | Logging level (debug/info/warn/error)        | No       |
| `LOG_FORMAT`         | Logging format (json/text)                   | No       |
| `SENTRY_DSN`         | Sentry DSN for error tracking                | No       |
| `ENVIRONMENT`        | Runtime environment (development/production) | No       |

*Required when `KAFKA_ENABLED=true`

### Sample Configuration

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_URL=postgres://user:pass@localhost/vpnbot?sslmode=disable
KAFKA_ENABLED=true
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=arcanus-events
LOG_LEVEL=info
ENVIRONMENT=development
```

## Development

### Prerequisites

- Go 1.25+
- PostgreSQL 15+
- Kafka (optional, mock publisher used when disabled)

### Setup

```bash
git clone <repository>
cd arcanus-vpn-telegram-bot
go mod download
cp env.example .env
# Configure .env file
go run cmd/bot/main.go
```

### Testing

```bash
go test ./...
go test ./... -coverprofile=coverage.out
```

### Continuous Integration

This project uses GitHub Actions for automated testing and building. The workflow includes:

- **Multi-version testing**: Tests against Go 1.21.x and 1.22.x
- **Database integration**: Uses PostgreSQL service container for integration tests
- **Code quality**: Automated linting with golangci-lint
- **Security scanning**: Vulnerability checks with govulncheck
- **Coverage reporting**: Generates HTML coverage reports
- **Binary artifacts**: Builds and uploads binaries for releases

The workflow runs on:
- Push to `main` and `develop` branches
- Pull requests to `main` branch

See [`.github/workflows/go.yml`](.github/workflows/go.yml) for the complete configuration.

### Architecture Principles

- **Clean Architecture** - Domain-driven design with dependency inversion
- **Event Sourcing** - All state changes captured in immutable event log  
- **SOLID Principles** - Maintainable, testable, extensible codebase
- **Dependency Injection** - Uber FX for component wiring and lifecycle