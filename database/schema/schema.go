package schema

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

/* AdminOnly is a marker struct to indicate that a field is admin only */
type AdminOnly struct {
	adminOnly string `gorm:"-:all"` // makes it so you can only access this field as admin (logged in as admin user)
}

type Hidden struct {
	Hidden string `gorm:"-:all"` // hidden from metadata
}

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewV4()
	return
}

type User struct {
	BaseModel
	AdminOnly
	Fullname string `gorm:"type:varchar(100);not null"`
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`

	Permissions uint64 `gorm:"not null;default:1;"`

	// adminOnly string `gorm:"-:all"` // makes it so you can only access this field as admin (logged in as admin user)
}

type UserToken struct {
	BaseModel
	Hidden
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Token     string    `gorm:"type:char(64);not null;uniqueIndex"`
	ExpiresAt int64     `gorm:"not null;index"`

	// Hidden string `gorm:"-:all"` // hidden from metadata
}

var DefaultSchema = []interface{}{
	&User{},
	&UserToken{},
}
