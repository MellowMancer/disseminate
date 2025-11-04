# Disseminate

A multi-platform social media management tool that allows you to post content to multiple platforms from a single dashboard.

## Roadmap

| Platform | Status | Notes |
|----------|--------|-------|
| **Twitter / X** | ✅ Completed | Full OAuth integration |
| **Instagram** | ✅ Completed* | Reel Posting Pending (Requires Facebook Integration) |
| **Facebook** | ⏳ Pending | Planned |
| **Bluesky** | 🚧 Ongoing | In development |
| **Mastodon** | ⏳ Pending | Planned |
| **Artstation** | ⏳ Pending | Planned |
| **YouTube** | ⏳ Pending | Planned |

*Instagram implementation may have some limitations

## Tech Stack

### Backend
- **Go 1.25+** - Main backend language
- **Echo** - Web framework
- **OAuth 1.0/2.0** - Social media authentication
- **Supabase** - Database
- **Cloudflare R2** - Media storage ( Supabase Free allowed only for 50MB )

### Frontend
- **React** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **shadcn/ui** - Component library

## Project Structure

```
disseminate/
├── backend/              # Go backend
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── repositories/     # Data access layer
│   ├── models/           # Data models
│   ├── middlewares/      # HTTP middlewares
│   └── routes/           # Route definitions
├── frontend/             # React frontend
│   └── src/
│       ├── components/   # React components
│       ├── pages/        # Page components
│       ├── context/      # React contexts
│       ├── lib/          # Utilities
│       └── types/        # TypeScript types
└── main.go               # Application entry point
```

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 18+
- Docker (optional)

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Twitter OAuth
TWITTER_CONSUMER_KEY=
TWITTER_CONSUMER_SECRET=
TWITTER_CALLBACK_URL=

# Instagram OAuth
INSTAGRAM_CLIENT_ID=
INSTAGRAM_CLIENT_SECRET=
INSTAGRAM_REDIRECT_URL=

# Supabase
SUPABASE_URL=
SUPABASE_KEY=

# Cloudflare R2
CLOUDFLARE_ACCOUNT_ID=
CLOUDFLARE_S3_API_URL=
CLOUDFLARE_TOKEN=
CLOUDFLARE_S3_ACCESS_KEY_ID=
CLOUDFLARE_S3_SECRET_ACCESS_KEY=

# JWT & Sessions
JWT_SECRET=
SESSION_SECRET=

# App Environment
APP_ENV=development
```

### Running Locally

1. **Backend**:
   ```bash
   go mod download
   air
   ```

2. **Frontend**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

### Docker

```bash
docker-compose up
```

## Development

### Backend Structure

- **Handlers**: Handle HTTP requests and responses
- **Services**: Contain business logic and orchestrate repository calls
- **Repositories**: Abstract data access (Supabase, Cloudflare, etc.)
- **Middleware**: JWT validation, request logging, etc.

### Frontend Structure

- **Components**: Reusable UI components (shadcn/ui based)
- **Pages**: Top-level page components
- **Context**: Global state management (Auth, Theme)
- **Types**: TypeScript type definitions
