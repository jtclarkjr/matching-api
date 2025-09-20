# Matching API

A complete Go-based REST API for a matching application built with Chi router, featuring real-time chat via WebSockets, sophisticated matching algorithms, and comprehensive user management.

## Features

### Core Services

- **Authentication Service**: JWT-based auth with refresh tokens
- **User Service**: Profile management, photos, preferences
- **Matching Service**: Swipe mechanics, compatibility scoring, discovery
- **Chat Service**: Real-time messaging with WebSocket support
- **Notification Service**: Push notifications (FCM/APNS), in-app, email
- **Image Service**: Photo upload, processing, storage integration
- **Analytics Service**: User behavior tracking, metrics, dashboard
- **Database Layer**: PostgreSQL with migrations

### Key Features

- **Smart Matching Algorithm**: Location-based, age preferences, compatibility scoring
- **Real-time Chat**: WebSocket connections for instant messaging
- **Typing Indicators**: See when matches are typing
- **Online Status**: Real-time online/offline status
- **Photo Management**: Upload, organize, and delete photos
- **Preference System**: Detailed matching preferences
- **Security**: JWT authentication, password hashing, input validation

## API Architecture

```
matching-api/
├── cmd/server/          # Application entry point
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # Custom middleware
│   ├── models/          # Data models
│   └── services/        # Business logic (planned)
├── pkg/
│   ├── auth/           # JWT utilities
│   ├── services/       # Redis and other services
│   └── utils/          # Helper functions
└── configs/            # Environment configurations (dev, prod, test)
```

## Quick Start

### 1. Install Dependencies

```bash
# Clone and navigate to project
cd matching-api

# Install Go dependencies
go mod tidy

# Install Swagger CLI tool
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. Configure Services (Optional)

```bash
# For S3 image storage (see docs/S3_IMAGE_SERVICE.md)
export AWS_S3_BUCKET=your-bucket-name
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
```

### 3. Generate Documentation

```bash
# Generate Swagger documentation
swag init -g cmd/server/main.go
```

### 4. Run the Server

```bash
# Start the server
go run cmd/server/main.go
```

The server starts on `http://localhost:8080`

### 5. Access Documentation

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

### 6. Test the API

#### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

## Docker Setup (Recommended)

The easiest way to run the matching API is using Docker Compose, which sets up the entire stack including PostgreSQL, Redis, and the API.

### Prerequisites

- Docker and Docker Compose installed
- Git (to clone the repository)

### Quick Start with Docker

1. **Clone and setup**

```bash
cd matching-api

# Copy environment template
cp .env.example .env

# Edit .env file with your configuration
nano .env
```

2. **Start all services**

```bash
# Build and start all services
make dev

# Or manually:
docker compose up -d --build
```

3. **Check service status**

```bash
# View logs
make logs

# Check health
curl http://localhost:8080/api/v1/health
```

4. **Access services**

- **API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **pgAdmin** (optional): http://localhost:5050
- **Redis Commander** (optional): http://localhost:8081

### Docker Commands

```bash
# Start with admin tools (pgAdmin + Redis Commander)
make tools

# View logs from specific service
docker compose logs -f app
docker compose logs -f redis
docker compose logs -f postgres

# Connect to services
make redis-cli    # Connect to Redis
make psql        # Connect to PostgreSQL
make shell       # Get shell in app container

# Stop services
make down

# Clean up everything (containers, volumes, images)
make clean
```

### Environment Configuration

The `.env` file controls all service configuration:

```bash
# Database
DB_NAME=matching_db
DB_USER=postgres
DB_PASSWORD=your_secure_password

# Redis
REDIS_PASSWORD=your_redis_password

# JWT Secret (change in production)
JWT_SECRET=your-super-secret-jwt-key-min-32-chars

# Optional AWS S3 (leave empty if not using)
AWS_S3_BUCKET=
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
```

#### Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "age": 28,
    "gender": "male"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

#### Get Profile (requires auth token)

```bash
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Docs

Can run OpenAPI swagger using `/swagger/index.html`

Generate a new yaml using `swag init`

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token

### User Management

- `GET /api/v1/users/profile` - Get current user profile
- `PUT /api/v1/users/profile` - Update user profile
- `POST /api/v1/users/photos` - Upload photo
- `DELETE /api/v1/users/photos/{photoID}` - Delete photo
- `GET /api/v1/users/preferences` - Get matching preferences
- `PUT /api/v1/users/preferences` - Update preferences

### Images (S3-powered)

- `GET /api/v1/images` - List user images
- `POST /api/v1/images/upload` - Upload image (multipart)
- `POST /api/v1/images/upload-base64` - Upload image (base64)
- `POST /api/v1/images/presigned-upload` - Generate presigned URL
- `GET /api/v1/images/download/{imageKey}` - Download image
- `DELETE /api/v1/images/{imageKey}` - Delete image

### Matching

- `POST /api/v1/matches/swipe` - Swipe on a user (like/pass/super like)
- `GET /api/v1/matches` - Get current matches
- `GET /api/v1/matches/potential` - Get potential matches
- `DELETE /api/v1/matches/{matchID}` - Remove a match

### Chat

- `GET /api/v1/chats` - Get chat conversations
- `GET /api/v1/chats/{chatID}/messages` - Get chat messages
- `POST /api/v1/chats/{chatID}/messages` - Send message
- `GET /api/v1/ws?user_id=USER_ID` - WebSocket connection for real-time features

### Notifications

- `GET /api/v1/notifications` - Get user notifications
- `PUT /api/v1/notifications/{notificationID}/read` - Mark notification as read
- `PUT /api/v1/notifications/read-all` - Mark all notifications as read
- `GET /api/v1/notifications/unread-count` - Get unread count
- `GET /api/v1/notifications/preferences` - Get notification preferences
- `PUT /api/v1/notifications/preferences` - Update notification preferences
- `POST /api/v1/notifications/devices` - Register device token
- `DELETE /api/v1/notifications/devices/{tokenID}` - Unregister device

### Analytics (Admin)

- `GET /api/v1/analytics/dashboard` - Dashboard metrics
- `GET /api/v1/analytics/users/{userID}` - User-specific metrics
- `GET /api/v1/analytics/funnel` - Conversion funnel analysis
- `POST /api/v1/analytics/track` - Track custom events

## WebSocket Usage

Connect to WebSocket for real-time features:

```javascript
const ws = new WebSocket("ws://localhost:8080/api/v1/ws?user_id=your-user-id");

// Listen for messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};

// Send typing indicator
ws.send(
  JSON.stringify({
    type: "typing",
    data: {
      chat_id: "chat-123",
      is_typing: true,
    },
  })
);
```

## Matching Algorithm

The API includes a sophisticated matching algorithm that considers:

1. **User Preferences**: Age range, distance, gender preferences
2. **Compatibility Score**: Based on multiple factors:
   - Age similarity (closer ages = higher score)
   - Location proximity (closer = higher score)
   - Recent activity (active users = higher score)
   - Profile completeness (photos, bio = higher score)
3. **Mutual Interest**: Both users must meet each other's criteria
4. **Geographic Distance**: Uses Haversine formula for accurate distance calculation

## Redis Caching & Sessions

The API uses Redis for high-performance caching and session management:

### Caching Features

- **User Profile Caching**: Frequently accessed user data with TTL
- **Match Caching**: Potential matches and existing matches
- **Session Management**: JWT token storage and validation
- **Rate Limiting**: Request throttling per user/IP

### Cache Keys

- `user:{userID}` - User profile data
- `matches:{userID}` - User's current matches
- `potential_matches:{userID}` - Suggested matches
- `session:{sessionID}` - User session data
- `rate_limit:{identifier}` - Rate limiting counters

### Performance Benefits

- **Reduced Database Load**: Frequently accessed data served from memory
- **Faster Response Times**: Sub-millisecond cache lookups
- **Scalable Sessions**: Distributed session storage
- **Automatic Expiration**: TTL-based cache invalidation

## Configuration Management

The API includes comprehensive configuration management via the `configs/` directory:

### Configuration Files

- **`app.yaml`**: Base configuration with all options documented
- **`development.yaml`**: Development-optimized settings
- **`production.yaml`**: Production-ready secure configuration
- **`testing.yaml`**: Test environment isolation settings

### Current Implementation

The API uses **environment variables** for configuration (12-factor app compliance):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=matching_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Future Enhancement

YAML files provide structure for future **configuration loading** and **multi-environment** deployment.

## Code Quality & Security

### Security Features

- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt for secure password storage
- **Input Validation**: Comprehensive request validation
- **CORS Protection**: Configurable cross-origin policies
- **Rate Limiting**: Redis-backed request throttling
- **Session Security**: Secure cookie handling with Chi sessions

### Code Quality

- **Linting**: golangci-lint compliant with zero issues
- **Error Handling**: Comprehensive error checking and proper error wrapping
- **Type Safety**: No `any` types, strict type checking throughout
- **Chi Ecosystem**: Using official Chi packages for reliability
- **Production Patterns**: Follows Go and Chi best practices

## Environment Variables

```bash
JWT_SECRET=your-super-secret-jwt-key
PORT=8080
DB_URL=your-database-url  # When database is added
```

## Current Status

### Items

- **Core API structure** with Chi router
- **Authentication system** with JWT and refresh tokens
- **User profile management** with photos and preferences
- **Sophisticated matching algorithm** with location-based scoring
- **Real-time chat** with WebSocket support
- **Push notification system** (FCM/APNS integration)
- **In-app notification** management
- **Email notification** system
- **Analytics service** with comprehensive metrics
- **User behavior tracking** with event system
- **Admin dashboard** with detailed insights
- **Image service** models and processing pipeline
- **PostgreSQL database** with full schema and migrations
- **Comprehensive error handling** and validation
- **Production-ready architecture**
- Input validation

### Ready for Enhancement

- **Production deployment** (Docker, Kubernetes)
- **Redis caching** for performance optimization
- **AWS S3/CloudFront** integration for image storage
- **Rate limiting** middleware implementation
- **Email verification** workflow
- **Payment processing** (Stripe integration)
- **Advanced ML** matching algorithms
- **Content moderation** and safety features
- Admin dashboard

## Tech Stack

- **Language**: Go 1.24+
- **Router**: Chi v5
- **Database**: PostgreSQL 15+ with migrations
- **Caching**: Redis 7+ for sessions and performance
- **Authentication**: JWT with golang-jwt/jwt
- **WebSockets**: Gorilla WebSocket
- **Validation**: go-playground/validator
- **Security**: bcrypt for passwords
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean architecture with dependency injection

## Development Commands

```bash
# Code Quality
go fmt ./...           # Format code
golangci-lint run      # Run linter (zero issues!)
go test -v ./...       # Run tests

# Generate/update Swagger documentation
swag init -g cmd/server/main.go

# Local Development
go run cmd/server/main.go    # Run locally
air                          # Hot reload (install: go install github.com/cosmtrek/air@latest)

# Docker Development
make dev              # Build and start all services
make tools            # Start with admin tools (pgAdmin + Redis Commander)
make logs             # View all logs
make shell            # Get shell in app container
make redis-cli        # Connect to Redis CLI
make psql             # Connect to PostgreSQL
make clean            # Clean up everything

# Production Build
go build -o matching-api cmd/server/main.go
docker compose build  # Build Docker image
```

## Production Readiness Checklist

- [x] **API documentation** (OpenAPI/Swagger)
- [x] **Database layer** (PostgreSQL with migrations)
- [x] **Redis for caching and sessions**
- [x] **Docker containerization**
- [ ] CI/CD pipeline
- [ ] Monitoring and logging
- [ ] Load testing
- [ ] Security audit

This API provides a solid foundation for a production matching app with room for horizontal scaling and feature expansion.

## License

MIT License - See LICENSE file for details.
