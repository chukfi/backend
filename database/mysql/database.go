package database

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"

	defaultSchema "github.com/chukfi/backend/database/schema"
	"github.com/chukfi/backend/src/lib/permissions"
	"github.com/chukfi/backend/src/lib/schemaregistry"
	"gorm.io/gorm"
)

// DB is the global database connection (MYSQL, *gorm.DB)
var DB *gorm.DB

// The database schema to be migrated, can be appended to by using InitDatabase(schema []interface{})
var Schema *[]interface{}

/*
InitDatabase initializes the database connection and migrates the provided schemas.
It also creates a default admin user if it does not already exist.

Using: MYSQL, requires DATABASE_DSN in .env file
*/
func InitDatabase(schema []interface{}) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_DSN")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		panic("failed to connect database:" + err.Error())
	}

	Schema := append(schema, defaultSchema.DefaultSchema...)

	schemaregistry.RegisterSchemas(Schema)

	db.AutoMigrate(Schema...)

	// print every schema that exists
	for _, s := range Schema {
		fmt.Printf("Database Schema: %T\n", s)
	}

	// create base user

	basePassword, err := bcrypt.GenerateFromPassword([]byte("chukfi123"), bcrypt.DefaultCost)
	if err != nil {
		panic("failed to hash base user password:" + err.Error())
	}

	user := defaultSchema.User{
		Fullname:    "Chukfi Admin",
		Password:    string(basePassword),
		Email:       "admin@nativeconsult.io",
		Permissions: uint64(permissions.Admin),
	}

	// check if user exists
	var existingUser defaultSchema.User
	result := db.Where("email = ?", user.Email).First(&existingUser)

	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		err = gorm.G[defaultSchema.User](db).Create(context.Background(), &user)

		if err != nil {
			panic("failed to create base user:" + err.Error())
		}
	}

	// setup :)
	DB = db

	if err := permissions.InitPermissions(db); err != nil {
		panic("failed to initialize permissions: " + err.Error())
	}
}
