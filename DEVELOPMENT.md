# Development Guide

## Prerequisites

Before you can run Chukfi CMS, ensure you have the following installed:

- **Node.js 18+** - For the frontend
- **Go 1.21+** - For the backend API
- **Git** - For version control

## Project Structure

```
chukfi-cms/
├── .github/
│   └── copilot-instructions.md    # Copilot workspace configuration
├── .vscode/
│   └── tasks.json                 # VS Code tasks for development
├── frontend/                      # Astro + React frontend
│   ├── src/
│   │   ├── components/           # React components
│   │   ├── layouts/              # Astro layouts
│   │   ├── pages/                # Astro pages & routes
│   │   ├── styles/               # Global styles
│   │   ├── lib/                  # Utilities
│   │   └── types/                # TypeScript types
│   ├── public/                   # Static assets
│   ├── astro.config.mjs          # Astro configuration
│   ├── tailwind.config.mjs       # Tailwind CSS config
│   └── package.json
├── backend/                      # Go HTTP API
│   ├── cmd/
│   │   ├── server/               # Main server entry point
│   │   └── migrate/              # Database migration tool
│   ├── internal/
│   │   ├── auth/                 # Authentication & JWT
│   │   ├── db/                   # Database connection
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/           # HTTP middleware
│   │   └── models/               # Data models
│   ├── migrations/               # SQL migration files
│   ├── go.mod                    # Go module definition
│   └── .air.toml                 # Hot reload configuration
├── shared/                       # Shared types between frontend/backend
│   └── types.ts
├── .env.example                  # Example environment variables
└── README.md
```

## Quick Start

### Using VS Code Tasks (Recommended)

1. Open the project in VS Code
2. Open Command Palette (`Ctrl+Shift+P`)
3. Run "Tasks: Run Task"
4. Select "Start Development Environment"

This will automatically:

- Run database migrations
- Start frontend development server (port 4321)
- Start backend API server (port 8080)

### Manual Setup

1. **Set up Environment**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Backend Setup**

   ```bash
   cd backend
   go mod tidy
   go run cmd/migrate/main.go up  # Run migrations
   go run cmd/server/main.go      # Start server
   ```

3. **Frontend Setup** (in a new terminal)
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Development Workflow

### Adding New Features

1. **Collections**: Create new content types

   - Backend: Add handlers in `internal/handlers/`
   - Frontend: Add pages in `src/pages/admin/`

2. **Database Changes**: Create migrations

   ```bash
   cd backend
   # Create new migration manually
   # Add files: migrations/XXX_migration_name.up.sql and migrations/XXX_migration_name.down.sql
   go run cmd/migrate/main.go up  # Apply migrations
   ```

3. **Frontend Components**: Add to `src/components/`
   - Use React for interactive components
   - Use Astro for static layouts and pages

### Available VS Code Tasks

- `Start Frontend Development` - Runs Astro dev server
- `Start Backend Development` - Runs Go server
- `Build Frontend` - Build production frontend
- `Build Backend` - Compile Go binary
- `Run Database Migrations` - Apply database migrations
- `Start Development Environment` - Start everything

## Environment Variables

Create a `.env` file in the root directory:

```env
# Database (SQLite - Pure Go implementation)
DATABASE_URL=sqlite://./data/chukfi.db
DB_DRIVER=sqlite

# JWT Authentication
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Server Configuration
PORT=8080
HOST=localhost
ENVIRONMENT=development

# Frontend API URL (for SSR)
API_URL=http://localhost:8080
```

## Testing the Setup

1. **Frontend**: http://localhost:4321

   - Should show the Chukfi CMS welcome page

2. **Backend Health Check**: http://localhost:8080/health

   - Should return `{"status":"ok","service":"chukfi-cms"}`

3. **Admin Dashboard**: http://localhost:4321/admin

   - Should show the admin dashboard

4. **Default Login**:
   - Email: `admin@chukfi.com`
   - Password: `admin123`

## Common Issues

### Go Not Found

- Install Go from https://golang.org/dl/
- Ensure Go is in your PATH

### Port Conflicts

- Frontend (4321): Change in `frontend/astro.config.mjs`
- Backend (8080): Change PORT in `.env`

### Database Connection

- Ensure SQLite database directory exists: `mkdir -p backend/data`
- Check database file: `ls -la backend/data/chukfi.db`

## Hot Reload

- **Frontend**: Astro provides built-in hot reload
- **Backend**: Manual restart required (or use a tool like Air)

## Building for Production

```bash
# Frontend
cd frontend
npm run build

# Backend
cd backend
go build -o bin/server cmd/server/main.go
```

## Database Management

### Migrations

- Create: Add files to `backend/migrations/` directory
- Apply: `go run cmd/migrate/main.go up`
- Rollback: `go run cmd/migrate/main.go down`

### Reset Database

```bash
rm -f backend/data/chukfi.db  # Remove database file
cd backend && go run cmd/migrate/main.go up  # Recreate
```

## API Documentation

The REST API follows these conventions:

- Base URL: `http://localhost:8080/api/v1`
- Authentication: Bearer token in Authorization header
- Response format: JSON
- Error format: `{"error": "code", "message": "description"}`

Key endpoints:

- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/me` - Current user
- `GET /api/v1/users` - List users (admin)
- `GET /api/v1/collections` - List collections
- `GET /api/v1/collections/{slug}/documents` - List documents

## Extending the CMS

### Custom Field Types

1. Update collection schema validation in `backend/internal/models/`
2. Add UI components in `frontend/src/components/form/`
3. Update form builder in `frontend/src/components/FormBuilder.tsx`

### Custom Hooks

Add lifecycle hooks in Go:

```go
// Register hooks for collections
hooks.RegisterHook("posts", "beforeCreate", func(data interface{}) error {
    // Custom logic
    return nil
})
```

### Custom Permissions

Extend the RBAC system by modifying:

- `backend/internal/auth/auth.go` - Permission checking
- `backend/internal/middleware/auth.go` - Route protection
- Frontend role management UI

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes following the coding standards
4. Test your changes
5. Submit a pull request

See the main README.md for full contribution guidelines.
