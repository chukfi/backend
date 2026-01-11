# Chukfi CMS Backend

A Go-based headless CMS backend framework built with GORM and Chi router. Provides automatic schema registration, authentication, permissions, and TypeScript type generation.

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Database](#database)
6. [Schema Definition](#schema-definition)
7. [Server & Routing](#server--routing)
8. [Authentication](#authentication)
9. [Permissions System](#permissions-system)
10. [API Endpoints](#api-endpoints)
11. [TypeScript Type Generation](#typescript-type-generation)
12. [Utilities](#utilities)
13. [Architecture](#architecture)

---

## Overview

Chukfi CMS is a modular backend framework that provides:

- **Automatic schema registration** with metadata extraction
- **Built-in authentication** with JWT-like token management
- **Bitwise permission system** with custom permission support
- **Dynamic collection endpoints** for CRUD operations
- **TypeScript type generation** from Go struct definitions
- **User caching** for performance optimization
- **Case-insensitive routing** with collection name preservation

### Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25+ |
| ORM | GORM |
| Router | Chi v5 |
| Database | MySQL/TiDB (primary), extensible |
| Auth | Bearer tokens + Cookies |

---

## Installation

### As a Go Module

```go
// go.mod
module your-project

replace github.com/chukfi/backend => /path/to/chukfi-cms/backend

go 1.25.3

require github.com/chukfi/backend v0.0.0
```

### Dependencies

The framework automatically includes:
- `gorm.io/gorm` - ORM
- `gorm.io/driver/mysql` - MySQL driver
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/joho/godotenv` - Environment variables
- `github.com/satori/go.uuid` - UUID generation
- `golang.org/x/crypto/bcrypt` - Password hashing

---

## Quick Start

```go
package main

import (
    database "github.com/chukfi/backend/database/mysql"
    "github.com/chukfi/backend/server/router"
    "github.com/chukfi/backend/server/serve"
)

type Post struct {
    schema.BaseModel
    Title string `gorm:"type:varchar(255);not null"`
    Body  string `gorm:"type:text;not null"`
}

func main() {
    customSchema := []interface{}{&Post{}}
    
    database.InitDatabase(customSchema)
    
    r := router.SetupRouter(database.DB)
    
    config := serve.NewServeConfig("3000", customSchema, database.DB, r)
    serve.Serve(config)
}
```

---

## Configuration

### Environment Variables

Create a `.env` file in your project root:

```env
DATABASE_DSN="user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

### Docker (TiDB Example)

```yaml
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

---

## Database

### Initialization

```go
import database "github.com/chukfi/backend/database/mysql"

customSchema := []interface{}{&MyModel{}, &AnotherModel{}}
database.InitDatabase(customSchema)

db := database.DB
```

The `InitDatabase` function:
1. Loads environment variables
2. Connects to the database
3. Registers all schemas (custom + default)
4. Auto-migrates all models
5. Creates a default admin user
6. Initializes the permission system

### Default Admin User

Automatically created on first run:
- **Email:** `admin@nativeconsult.io`
- **Password:** `chukfi123`
- **Permissions:** Administrator

### Database Helper

A fluent query builder wrapper around GORM:

```go
import helper "github.com/chukfi/backend/database/helper"

posts, err := helper.Get[Post](db).
    Where("title LIKE ?", "%test%").
    Order("created_at DESC").
    Limit(10).
    Find()

post, err := helper.Get[Post](db).
    Where("id = ?", id).
    First()

count, err := helper.Get[Post](db).
    Where("author_id = ?", authorID).
    Count()

err := helper.Get[Post](db).Create(&newPost)

err := helper.Get[Post](db).
    Where("id = ?", id).
    Updates(map[string]interface{}{"title": "New Title"})

helper.Paginate[Post](db, page, pageSize).Find()
```

**Available Methods:**
| Method | Description |
|--------|-------------|
| `Where`, `Or`, `Not` | Query conditions |
| `Select`, `Omit` | Column selection |
| `Order`, `Limit`, `Offset` | Pagination/sorting |
| `Group`, `Having` | Aggregation |
| `Joins`, `Preload` | Relationships |
| `First`, `Last`, `Take`, `Find` | Retrieval |
| `Count`, `Exists` | Counting |
| `Create`, `Save`, `Update`, `Updates`, `Delete` | Mutations |
| `FirstOrCreate`, `FirstOrInit` | Upsert operations |

---

## Schema Definition

### Base Model

All models should embed `BaseModel` for standard fields:

```go
import "github.com/chukfi/backend/database/schema"

type MyModel struct {
    schema.BaseModel
    Name string `gorm:"type:varchar(100);not null"`
}
```

`BaseModel` provides:
- `ID` - UUID primary key (auto-generated)
- `CreatedAt` - Timestamp
- `UpdatedAt` - Timestamp  
- `DeletedAt` - Soft delete support

### Marker Structs

#### AdminOnly

Models embedding `AdminOnly` require admin permissions for API access:

```go
type SecretConfig struct {
    schema.BaseModel
    schema.AdminOnly
    Key   string `gorm:"type:varchar(100);not null"`
    Value string `gorm:"type:text"`
}
```

#### Hidden

Models embedding `Hidden` are excluded from schema registration and API exposure
They can however, be used with gorm / the database helper to save data you do not want edited within the CMS

```go
type InternalToken struct {
    schema.BaseModel
    schema.Hidden
    Token string `gorm:"type:char(64)"`
}
```

### Default Models

**User**
```go
type User struct {
    BaseModel
    AdminOnly
    Fullname    string `gorm:"type:varchar(100);not null"`
    Email       string `gorm:"type:varchar(100);uniqueIndex;not null"`
    Password    string `gorm:"type:varchar(255);not null"`
    Permissions uint64 `gorm:"not null;default:1"`
}
```

**UserToken**
```go
type UserToken struct {
    BaseModel
    Hidden
    UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
    Token     string    `gorm:"type:char(64);not null;uniqueIndex"`
    ExpiresAt int64     `gorm:"not null;index"`
}
```

---

## Server & Routing

### Setup

```go
import (
    "github.com/chukfi/backend/server/router"
    "github.com/chukfi/backend/server/serve"
)

r := router.SetupRouter(database.DB)

r.Get("/custom", customHandler)

config := serve.NewServeConfig("3000", schema, database.DB, r)
serve.Serve(config)
```

### Default Middleware

Applied automatically by `SetupRouter`:

1. **Logger** - Request logging
2. **CaseSensitiveMiddleware** - Lowercases paths (preserves collection names)
3. **SaveAuthTokenMiddleware** - Extracts auth token from cookies/headers

### HTTP Responder

```go
import "github.com/chukfi/backend/src/httpresponder"

httpresponder.SendNormalResponse(w, r, data)

httpresponder.SendErrorResponse(w, r, "Error message", http.StatusBadRequest)
```

---

## Authentication

### Login Flow

1. Client POSTs to `/admin/auth/login` with email/password
2. Server validates credentials via bcrypt
3. Server creates `UserToken` with 24-hour expiry
4. Token returned in response + set as `chukfi_auth_token` cookie

### Token Usage

**Cookie (automatic):**
```
Cookie: chukfi_auth_token=<token>
```

**Header:**
```
Authorization: Bearer <token>
```

### Auth Middleware

```go
r.Route("/protected", func(r chi.Router) {
    r.Use(router.AuthMiddlewareWithDatabase(database.DB))
    
    r.Get("/resource", handler)
})
```

### User Retrieval

```go
userID := router.GetUserIDFromRequest(r)

user, err := router.GetUserFromRequest(r, database.DB)
```

---

## Permissions System

### Built-in Permissions

| Permission | Bit | Description |
|------------|-----|-------------|
| `ViewDashboard` | 0 | Access dashboard |
| `ViewModels` | 1 | View schema metadata |
| `ViewUsers` | 2 | View user list |
| `ManageUsers` | 3 | Create/edit/delete users |
| `ManageModels` | 4 | Full model CRUD access |
| `Administrator` | 5 | Full system access |

### Permission Groups

```go
permissions.BasicUser // ViewDashboard | ViewModels | ViewUsers
permissions.Admin     // Administrator (bypasses all checks)
```

### Custom Permissions

```go
import "github.com/chukfi/backend/src/lib/permissions"

perm, err := permissions.RegisterPermission("ViewPosts")

err := permissions.UnregisterPermission("ViewPosts")
```

Custom permissions are persisted in the `custom_permissions` table.

### Permission Checking

```go
hasAccess := router.RequestRequiresPermission(r, db, permissions.ManageModels)

r.Use(router.RoutesRequiresPermission(db, permissions.Admin))

if permissions.HasPermission(permissions.Permission(user.Permissions), permissions.ViewModels) {
    // allowed
}
```

### Permission Utilities

```go
names := permissions.PermissionsToStrings(permissions.Permission(user.Permissions))

combined := permissions.StringsToPermission([]string{"ViewDashboard", "ViewModels"})

all := permissions.AllPermissionsAsStrings()
```

---

## API Endpoints

### Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/admin/auth/login` | No | Login, returns token |
| GET | `/admin/auth/me` | Yes | Get current user info |

**Login Request:**
```json
{
  "email": "admin@nativeconsult.io",
  "password": "chukfi123"
}
```

**Login Response:**
```json
{
  "authToken": "uuid-token",
  "expiresAt": 1234567890,
  "user": {
    "id": "uuid",
    "fullname": "Chukfi Admin",
    "email": "admin@nativeconsult.io",
    "permissions": ["Administrator"]
  },
  "success": true
}
```

### Collections

| Method | Endpoint | Auth | Permission | Description |
|--------|----------|------|------------|-------------|
| GET | `/admin/collection/all` | Yes | ViewModels | List all schemas |
| GET | `/admin/collection/{name}/metadata` | Yes | ViewModels | Get schema metadata |
| GET | `/admin/collection/{name}/get` | Conditional | — | Get all records |
| POST | `/admin/collection/{name}/create` | Yes | — | Create record |

**Collection Names:** Support both singular and plural forms (`post` or `posts`).

**AdminOnly Collections:** Require `ManageModels` permission for access.

**Metadata Response:**
```json
{
  "TableName": "posts",
  "AdminOnly": false,
  "Fields": [
    {
      "Name": "ID",
      "Type": "uuid.UUID",
      "GormTag": "type:char(36);primaryKey",
      "Required": true,
      "PrimaryKey": true
    }
  ]
}
```

---

## TypeScript Type Generation

### CLI Usage

```bash
./chukfi-cms generate-types --schema=./schema.go

./chukfi-cms generate-types \
  --schema=./schema.go \
  --output=./types/api.ts \
  --dsn="user:password@tcp(127.0.0.1:3306)/dbname"
```

**Options:**
| Flag | Required | Description |
|------|----------|-------------|
| `--schema` | Yes | Path to Go schema file |
| `--output` | No | Output path (default: `./cms.types.ts`) |
| `--dsn` | No | Database DSN (uses env if not set) |
| `--database` | No | Database type (`mysql`/`postgres`) |

### Programmatic Usage

```go
import generate_types "github.com/chukfi/backend/cmd/generate-types"

config := &generate_types.GenerateTypesConfig{
    Schema:   customSchema,
    Database: database.DB,
}
bytes := generate_types.GenerateTypescriptTypes(config)
```

### Generated Output Example

```typescript
export interface Post {
  ID: string;
  CreatedAt?: string;
  UpdatedAt?: string;
  DeletedAt?: string;
  Title?: string;
  Body?: string;
  AuthorID?: string;
}
```

### Type Mapping

| Go Type | TypeScript Type |
|---------|-----------------|
| `string`, `uuid.UUID` | `string` |
| `int`, `uint`, `float` | `number` |
| `bool` | `boolean` |
| `time.Time` | `string` or `Date` |
| `[]T` | `T[]` |
| `map[K]V` | `Record<string, any>` |

---

## Utilities

### User Cache

In-memory cache with TTL (default: 5 minutes):

```go
import usercache "github.com/chukfi/backend/src/lib/cache/user"

user, found := usercache.UserCacheInstance.Get(userID)

usercache.UserCacheInstance.Set(userID, user)

usercache.UserCacheInstance.Delete(userID)

usercache.UserCacheInstance.Clear()
```

### Schema Registry

```go
import "github.com/chukfi/backend/src/lib/schemaregistry"

tableName, exists := schemaregistry.ResolveTableName("post")

metadata, exists := schemaregistry.GetMetadata("posts")

fields, exists := schemaregistry.GetFields("posts")

required := schemaregistry.GetRequiredFields("posts")

isAdmin := schemaregistry.IsAdminOnly("users")

missing, unknown := schemaregistry.ValidateBody("posts", bodyMap)
```

### Database Detection

```go
import "github.com/chukfi/backend/src/lib/detection"

dbType := detection.DetectDatabaseType(dsn)
// Returns: MySQL, PostgreSQL, SQLite, or Unknown
```

### AST Parser

For parsing Go struct files:

```go
import "github.com/chukfi/backend/src/lib/astparser"

structs, err := astparser.ParseSchemaFile("./schema.go")

typescript := astparser.GenerateTypescriptFromParsed(structs)
```

---

## Architecture

```
backend/
├── main.go                          # CLI entry point
├── go.mod                           # Module definition
├── cmd/
│   └── generate-types/
│       └── main.go                  # Type generation logic
├── database/
│   ├── helper/
│   │   └── helper.go                # Query builder
│   ├── mysql/
│   │   └── database.go              # Database initialization
│   └── schema/
│       └── schema.go                # Base models (User, UserToken)
├── internal/
│   └── cli/
│       └── generate-types/
│           └── generate-types.go    # CLI handler
├── server/
│   ├── router/
│   │   └── router.go                # Chi router setup + handlers
│   └── serve/
│       └── serve.go                 # Server configuration
└── src/
    ├── chumiddleware/
    │   └── middleware.go            # Custom middleware
    ├── httpresponder/
    │   └── httpresponder.go         # HTTP response helpers
    └── lib/
        ├── astparser/
        │   └── parser.go            # Go AST parsing
        ├── cache/
        │   └── user/
        │       └── usercache.go     # User caching
        ├── detection/
        │   └── database.go          # DB type detection
        ├── permissions/
        │   ├── model.go             # CustomPermission model
        │   └── permission.go        # Permission system
        └── schemaregistry/
            └── registry.go          # Schema metadata registry
```

---

## Example Project Structure

```
your-project/
├── .env
├── go.mod
├── main.go
├── schema.go
├── docker-compose.yml
└── cms.types.ts          # Generated
```

**main.go:**
```go
package main

import (
    database "github.com/chukfi/backend/database/mysql"
    "github.com/chukfi/backend/server/router"
    "github.com/chukfi/backend/server/serve"
)

func main() {
    customSchema := []interface{}{&Post{}, &Category{}}
    database.InitDatabase(customSchema)
    
    r := router.SetupRouter(database.DB)
    
    // Add custom routes here
    
    config := serve.NewServeConfig("3000", customSchema, database.DB, r)
    serve.Serve(config)
}
```

**schema.go:**
```go
package main

import "github.com/chukfi/backend/database/schema"

type Post struct {
    schema.BaseModel
    Title    string `gorm:"type:varchar(255);not null"`
    Body     string `gorm:"type:text;not null"`
    AuthorID string `gorm:"type:char(36);index"`
}

type Category struct {
    schema.BaseModel
    schema.AdminOnly
    Name string `gorm:"type:varchar(100);not null;uniqueIndex"`
}
```
