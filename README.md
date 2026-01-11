<div align="center">
  <img src="./assets/branding/logo.jpg" alt="Chukfi CMS Logo" width="200" height="200" />
  
  # Chukfi CMS
  
  **Chukfi** (chook-fee) is the Choctaw word for **rabbit**, a symbol of speed, agility, and quick thinking.<br>
  Chukfi CMS embraces those qualities by providing a fast, modern, open-source headless CMS built with **Go**.
</div>

<div align="center">

[![Release](https://img.shields.io/github/v/release/chukfi/backend?include_prereleases&style=flat-square)](https://github.com/chukfi/backend/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://go.dev/)

</div>

## Overview

Chukfi is a Go library for building content management systems. It provides authentication, permissions, schema registration, and a REST API out of the box. You define your schemas and Chukfi handles the rest.

## Installation

```bash
go get github.com/chukfi/backend
```

Or install the CLI globally:

```bash
go install github.com/chukfi/backend/cmd/chukfi@latest
```

## Quick Start

### 1. Set Up Database

Chukfi uses MySQL/TiDB. The easiest way to get started is with Docker:

```yaml
# docker-compose.yml
version: '3.8'

services:
  tidb:
    image: pingcap/tidb:latest
    ports:
      - "4002:4000"
    command:
      - --store=unistore
      - --path=/tmp/tidb
    volumes:
      - ./docker-tidb:/tmp/tidb
```

```bash
docker-compose up -d
```

### 2. Create Your Project

```bash
mkdir my-cms && cd my-cms
go mod init my-cms
go get github.com/chukfi/backend
go install github.com/chukfi/backend/cmd/chukfi@latest
chukfi setup-frontend # optional --url={url} --directory={directory}
```

### 3. Define Your Schema

```go
// schema.go
package main

import "github.com/chukfi/backend/database/schema"

type Post struct {
    schema.BaseModel
    Type     string `gorm:"type:varchar(100);not null"`
    Body     string `gorm:"type:text;not null"`
    Title    string `gorm:"type:varchar(255);not null;index"`
    AuthorID string `gorm:"type:char(36);index"`
}

type HiddenModel struct {
    schema.BaseModel
    schema.Hidden
    Key        string `gorm:"type:char(64);not null;uniqueIndex"`
}

type APIKeys struct {
    schema.BaseModel
    schema.AdminOnly
    Key        string `gorm:"type:char(64);not null;uniqueIndex"`
    OwnerEmail string `gorm:"type:varchar(100);not null;index"`
    ExpiresAt  int64  `gorm:"not null;index"`
}

// or any other schemas, there are examples
```

Embed `schema.AdminOnly` in any model to restrict access to authenticated admin users.
Embed `schema.Hidden` in any model to hide it from any user as well as the dashboard.

### 4. Create Your Server

```go
// main.go
package main

import (
    "net/http"

    database "github.com/chukfi/backend/database/mysql"
    "github.com/chukfi/backend/server/router"
    "github.com/chukfi/backend/server/serve"
    "github.com/chukfi/backend/src/httpresponder"
    "github.com/chukfi/backend/src/lib/permissions"
    "github.com/go-chi/chi/v5"
    "gorm.io/gorm"
)

func main() {
    customSchema := []interface{}{
        &Post{},
        &APIKeys{},
        &HiddenModel{}
    }

    database.InitDatabase(customSchema)

    r := router.SetupRouter(database.DB, "./public") // or r := router.SetupRouter(database.DB) for no server.

    r.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
        posts, err := gorm.G[Post](database.DB).Find(r.Context())
        if err != nil {
            httpresponder.SendErrorResponse(w, r, err.Error(), 500)
            return
        }
        httpresponder.SendNormalResponse(w, r, posts)
    })

    r.Route("/api", func(r chi.Router) {
        r.Use(router.AuthMiddlewareWithDatabase(database.DB))

        r.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
            user, _ := router.GetUserFromRequest(r, database.DB)
            httpresponder.SendNormalResponse(w, r, user)
        })
    })

    serveConfig := serve.NewServeConfig("3000", []interface{}{}, database.DB, r)
    serve.Serve(serveConfig)
}
```

### 5. Configure Environment

```bash
# .env
DATABASE_DSN="root:@tcp(127.0.0.1:4002)/test?charset=utf8mb4&parseTime=True&loc=Local"
```

### 6. Run

```bash
go run .
```

## Features

### Schema System

All models should embed `schema.BaseModel` which provides:
- `ID` (UUID)
- `CreatedAt`
- `UpdatedAt`  
- `DeletedAt` (soft delete)

```go
type Product struct {
    schema.BaseModel
    Name  string `gorm:"type:varchar(255);not null"`
    Price int    `gorm:"not null"`
}
```

### Admin-Only Models

Embed `schema.AdminOnly` to restrict model access to authenticated users with admin permissions:

```go
type SecretConfig struct {
    schema.BaseModel
    schema.AdminOnly
    Key   string `gorm:"type:varchar(100)"`
    Value string `gorm:"type:text"`
}
```

### Authentication

Chukfi provides built-in auth endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/auth/login` | POST | Login with email/password, returns auth token |
| `/admin/auth/me` | GET | Get current user info |

Use `router.AuthMiddlewareWithDatabase(db)` to protect routes:

```go
r.Route("/protected", func(r chi.Router) {
    r.Use(router.AuthMiddlewareWithDatabase(database.DB))
    
    r.Get("/data", handler)
})
```

### Permissions

Register custom permissions and check them in handlers:

```go
viewPosts, _ := permissions.RegisterPermission("ViewPosts")

r.Get("/posts", func(w http.ResponseWriter, r *http.Request) {
    if !router.RequestRequiresPermission(r, database.DB, viewPosts) {
        httpresponder.SendErrorResponse(w, r, "Forbidden", 403)
        return
    }
    // ...
})
```

Or use middleware for entire route groups:

```go
r.Route("/admin-only", func(r chi.Router) {
    r.Use(router.RoutesRequiresPermission(database.DB, permissions.ManageModels))
    // all routes here require ManageModels permission
})
```

### Collection API

Chukfi automatically provides REST endpoints for registered schemas:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/collection/all` | GET | List all collections (requires auth) |
| `/admin/collection/{name}/get` | GET | Get all entries in collection |
| `/admin/collection/{name}/create` | POST | Create new entry (requires auth) |
| `/admin/collection/{name}/metadata` | GET | Get collection schema metadata |

### Database Helper

Use the typed query builder for cleaner database operations:

```go
import databasehelper "github.com/chukfi/backend/database/helper"

post, err := databasehelper.Get[Post](database.DB).
    Where("title = ?", "My Post").
    First()

posts, err := databasehelper.Get[Post](database.DB).
    Where("author_id = ?", userID).
    Order("created_at DESC").
    Limit(10).
    Find()
```

### TypeScript Type Generation

Generate TypeScript types from your Go schemas:

```go
import generate_types "github.com/chukfi/backend/cmd/generate-types"

generate_types.GenerateTypescriptTypes(&generate_types.GenerateTypesConfig{
    Schema:   customSchema,
    Database: database.DB,
})
```

Or use the CLI:

```bash
chukfi generate-types --schema=path/to/schema.go # optionally --dsn={dsn} --database={database provider (mysql/postgres)}
```

### Admin Frontend Serving

Serve a static admin frontend directory:

```go
r := router.SetupRouter(database.DB, "./public")
```

Build and download the admin frontend automatically:

```bash
chukfi setup-frontend # --url=https://github.com/your/frontend.git
```

## CLI Commands

```bash
chukfi generate-types    # Generate TypeScript types from database schema
chukfi setup-frontend     # Clone, build, and serve frontend
```

## License

MIT License - see [LICENSE](LICENSE) for details.
