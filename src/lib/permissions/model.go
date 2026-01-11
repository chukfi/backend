package permissions

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type CustomPermission struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	BitPosition uint      `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"type:varchar(255)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (base *CustomPermission) BeforeCreate(tx *gorm.DB) (err error) {
	base.ID = uuid.NewV4()
	return
}
