# Contributing to Chukfi CMS

Thank you for your interest in contributing to Chukfi CMS! This document provides guidelines and information for contributors.

## 🚀 Quick Start

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/Native-Consulting-Services/chukfi-cms.git
   cd chukfi-cms
   ```
3. **Set up development environment** following the [README.md](README.md#-fresh-installation-guide)
4. **Create a branch** for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## 📋 Development Workflow

### Prerequisites

- Go 1.21+
- Node.js 18+
- Git

### Setup

```bash
# Install dependencies
cd frontend && npm install && cd ../backend && go mod tidy && cd ..

# Start development environment
cd backend && mkdir -p data && go run cmd/migrate/main.go up
# Terminal 1: cd frontend && npm run dev
# Terminal 2: cd backend && go run cmd/server/main.go
```

### Testing

```bash
# Frontend
cd frontend && npm run build && npx astro check

# Backend
cd backend && go test ./... && go build cmd/server/main.go
```

## 🎯 Contribution Areas

### 🎨 Frontend (Astro + React + Tailwind)

- **Location**: `frontend/src/`
- **Components**: React components for admin UI
- **Pages**: Astro pages and layouts
- **Styles**: Tailwind CSS utilities

### 🔧 Backend (Go + SQLite)

- **Location**: `backend/internal/`
- **API**: REST endpoints and handlers
- **Auth**: JWT authentication and RBAC
- **Database**: SQLite operations and migrations

### 📚 Documentation

- **README.md**: Installation and usage guides
- **DEVELOPMENT.md**: Detailed development setup
- **API docs**: Endpoint documentation

### 🧪 Testing & CI

- **Frontend tests**: Component and build tests
- **Backend tests**: Unit and integration tests
- **GitHub Actions**: CI/CD workflows

## 📝 Coding Standards

### Go Code

- Follow standard Go conventions (`gofmt`, `golint`)
- Use meaningful variable and function names
- Add comments for complex logic
- Handle errors appropriately

### JavaScript/TypeScript

- Use TypeScript for type safety
- Follow Astro and React best practices
- Use Tailwind CSS for styling
- Keep components small and focused

### Git Commits

Use conventional commit messages:

```
type(scope): description

feat(auth): add password reset functionality
fix(ui): resolve mobile navigation issue
docs(readme): update installation instructions
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment details** (OS, Node.js, Go versions)
2. **Steps to reproduce** the issue
3. **Expected vs actual behavior**
4. **Error messages** or logs
5. **Screenshots** if UI-related

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

## ✨ Feature Requests

For new features:

1. **Check existing issues** to avoid duplicates
2. **Describe the problem** you're trying to solve
3. **Propose a solution** with implementation details
4. **Consider breaking changes** and backwards compatibility

Use the [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).

## 🔀 Pull Request Process

1. **Create a branch** from `main`:

   ```bash
   git checkout -b feature/your-feature
   ```

2. **Make your changes** following coding standards

3. **Test your changes**:

   ```bash
   # Frontend
   cd frontend && npm run build

   # Backend
   cd backend && go test ./... && go build cmd/server/main.go
   ```

4. **Update documentation** if needed

5. **Create pull request** with:

   - Clear title and description
   - Link to related issues
   - Screenshots for UI changes
   - Test results

6. **Respond to review feedback** promptly

### PR Requirements

- [ ] Code follows project conventions
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
- [ ] PR template filled out

## 🏗️ Architecture Overview

```
chukfi-cms/
├── frontend/           # Astro + React admin dashboard
│   ├── src/components/ # Reusable React components
│   ├── src/pages/      # Astro pages and routes
│   └── src/layouts/    # Page layouts
├── backend/            # Go HTTP API server
│   ├── cmd/           # Application entry points
│   ├── internal/      # Private application code
│   └── migrations/    # Database schema migrations
└── shared/            # Types shared between frontend/backend
```

### Key Design Principles

- **Zero external dependencies** for easy setup
- **Pure Go SQLite** for database (no CGO)
- **Collection-based** content management
- **JWT authentication** with RBAC
- **Modern web technologies**

## 🛡️ Security

- **Report security issues** privately via email
- **Don't commit** sensitive information (API keys, passwords)
- **Follow authentication** best practices
- **Validate user input** on both frontend and backend

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and community support
- **Documentation**: Check README.md and DEVELOPMENT.md first

## 📄 License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

## 🎉 Recognition

Contributors are recognized in:

- GitHub contributor graph
- Release notes (for significant contributions)
- Special thanks in documentation

Thank you for making Chukfi CMS better! 🚀
