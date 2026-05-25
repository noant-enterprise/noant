# Omagent Enterprise v2.0

## AI-Powered Customer Support Platform
Built in Nigeria. For the World.

## Tech Stack
- **Backend**: Go 1.22 + Gin Framework
- **Database**: TiDB Cloud (Distributed SQL)
- **Cache**: Redis (Upstash)
- **AI**: Groq Llama 3.3 with LangChain-style orchestration
- **Frontend**: Vanilla HTML5 + CSS3 + JavaScript (Enterprise UI)
- **Payments**: Polar (Naira support)

## Quick Start

```bash
# 1. Clone and enter backend
cd backend

# 2. Copy environment variables
cp .env.example .env
# Edit .env with your credentials

# 3. Download dependencies
go mod download

# 4. Run database migrations
# (Use TiDB Cloud console or mysql client)
mysql -h your-tidb-host -P 4000 -u your-user -p < migrations/001_init.sql

# 5. Start the server
go run main.go
```

## Features
- Multi-channel support (Telegram, WhatsApp, Web, Instagram, Messenger, Email)
- CSV-based AI training with instant deployment
- Knowledge gap detection with 100% accuracy
- Role-based team management (Owner, Admin, Agent)
- Real-time analytics and conversation monitoring
- Built-in Polar payment processing (Naira)
- JWT authentication with bcrypt password hashing
- Redis-based rate limiting and conversation memory
- Graceful shutdown and structured logging

## API Documentation
All endpoints are prefixed with `/api/v1`

### Authentication
- POST `/auth/register` - Create account
- POST `/auth/login` - User login
- POST `/auth/refresh` - Refresh token
- POST `/auth/change-password` - Change password

### Chats
- POST `/chats/direct-chat` - AI chat
- GET `/chats/conversations` - List conversations
- GET `/chats/conversations/:id` - Get conversation
- PUT `/chats/conversations/:id/takeover` - Human takeover

### Training
- GET `/training/categories` - List categories
- POST `/training/categories` - Create category
- POST `/training/bulk-qa` - Bulk import
- POST `/training/csv-upload` - CSV upload
- GET `/training/unknown-questions` - Knowledge gaps

### Analytics
- GET `/analytics/overview` - Dashboard metrics
- GET `/analytics/channels` - Channel distribution
- GET `/analytics/insights` - AI insights
- GET `/analytics/trends` - Historical trends

## Environment Variables
See `.env.example` for all required variables.

## License
MIT License — Built with ❤️ for Nigerian and African businesses.
