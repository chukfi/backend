package permissions

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

type Permission uint64

const (
	ViewDashboard Permission = 1 << iota
	ViewModels
	ViewUsers
	ManageUsers
	ManageModels
	Administrator

	maxBuiltinBit = iota
)

var (
	BasicUser = ViewDashboard | ViewModels | ViewUsers
	Admin     = Administrator
)

var (
	ErrMaxPermissionsReached = errors.New("maximum permissions reached (64)")
	ErrPermissionNotFound    = errors.New("permission not found")
)

type PermissionRegistry struct {
	mu               sync.RWMutex
	nameToPermission map[string]Permission
	permissionToName map[Permission]string
	nextBit          uint
	db               *gorm.DB
}

var registry = &PermissionRegistry{
	nameToPermission: make(map[string]Permission),
	permissionToName: make(map[Permission]string),
	nextBit:          maxBuiltinBit,
}

func init() {
	builtins := map[Permission]string{
		ViewDashboard: "ViewDashboard",
		ViewModels:    "ViewModels",
		ViewUsers:     "ViewUsers",
		ManageUsers:   "ManageUsers",
		ManageModels:  "ManageModels",
		Administrator: "Administrator",
	}

	for perm, name := range builtins {
		registry.permissionToName[perm] = name
		registry.nameToPermission[name] = perm
	}
}

// InitPermissions initializes the permission system with the given database connection.
func InitPermissions(db *gorm.DB) error {
	registry.mu.Lock()
	registry.db = db
	registry.mu.Unlock()

	db.AutoMigrate(&CustomPermission{})

	return LoadCustomPermissions(db)
}

// LoadCustomPermissions loads custom permissions from the database into the registry.
func LoadCustomPermissions(db *gorm.DB) error {
	var customPerms []CustomPermission
	if err := db.Order("bit_position asc").Find(&customPerms).Error; err != nil {
		return err
	}

	registry.mu.Lock()
	defer registry.mu.Unlock()

	for _, cp := range customPerms {
		perm := Permission(1 << cp.BitPosition)
		registry.nameToPermission[cp.Name] = perm
		registry.permissionToName[perm] = cp.Name

		if cp.BitPosition >= registry.nextBit {
			registry.nextBit = cp.BitPosition + 1
		}
	}

	return nil
}

// RegisterPermission registers a new custom permission with the given name.
func RegisterPermission(name string) (Permission, error) {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	foundperm, exists := registry.nameToPermission[name]

	if exists {
		return foundperm, nil
	}

	if registry.nextBit >= 64 {
		return 0, ErrMaxPermissionsReached
	}

	bitPos := registry.nextBit
	perm := Permission(1 << bitPos)
	registry.nextBit++

	registry.nameToPermission[name] = perm
	registry.permissionToName[perm] = name

	if registry.db != nil {
		customPerm := CustomPermission{
			Name:        name,
			BitPosition: bitPos,
		}
		if err := registry.db.Create(&customPerm).Error; err != nil {
			delete(registry.nameToPermission, name)
			delete(registry.permissionToName, perm)
			registry.nextBit--
			return 0, err
		}
	}

	return perm, nil
}

// UnregisterPermission removes a custom permission by name.
func UnregisterPermission(name string) error {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	perm, exists := registry.nameToPermission[name]
	if !exists {
		return ErrPermissionNotFound
	}

	for builtinPerm := range registry.permissionToName {
		if builtinPerm == perm && builtinPerm < (1<<maxBuiltinBit) {
			return errors.New("cannot unregister builtin permission")
		}
	}

	if registry.db != nil {
		if err := registry.db.Where("name = ?", name).Delete(&CustomPermission{}).Error; err != nil {
			return err
		}
	}

	delete(registry.nameToPermission, name)
	delete(registry.permissionToName, perm)

	return nil
}

// GetPermissionByName retrieves a permission by its name.
func GetPermissionByName(name string) (Permission, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	perm, ok := registry.nameToPermission[name]
	return perm, ok
}

// GetAllCustomPermissions retrieves all custom permissions from the database.
func GetAllCustomPermissions() []CustomPermission {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var perms []CustomPermission
	if registry.db != nil {
		registry.db.Find(&perms)
	}
	return perms
}

// HasPermission checks if the userPermissions include the requiredPermissions.
func HasPermission(userPermissions, requiredPermissions Permission) bool {
	if userPermissions&Administrator == Administrator {
		return true
	}
	return userPermissions&requiredPermissions == requiredPermissions
}

// AllPermissionsAsStrings returns a slice of all registered permission names.
func AllPermissionsAsStrings() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	names := make([]string, 0, len(registry.nameToPermission))
	for name := range registry.nameToPermission {
		names = append(names, name)
	}
	return names
}

// PermissionsToStrings converts a Permission bitmask to a slice of permission names.
func PermissionsToStrings(userPermissions Permission) []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var names []string
	for perm, name := range registry.permissionToName {
		if userPermissions&perm == perm {
			names = append(names, name)
		}
	}
	return names
}

// StringsToPermission converts a slice of permission names to a Permission bitmask.
func StringsToPermission(names []string) Permission {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var combined Permission
	for _, name := range names {
		if perm, ok := registry.nameToPermission[name]; ok {
			combined |= perm
		}
	}
	return combined
}

// PermissionToName retrieves the name of a permission given its bitmask.
func PermissionToName(permission Permission) string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	if name, ok := registry.permissionToName[permission]; ok {
		return name
	}
	return "Unknown"
}
