package model

// ─────────────────────────────────────────────
//  RBAC: User · Role · Permission
// ─────────────────────────────────────────────

// User adalah entitas pengguna sistem.
// Profil spesifik (guru/santri/wali) ada di domain akademik.
type User struct {
	Base
	Name     string `gorm:"type:varchar(255);not null"           json:"name"`
	Email    string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password string `gorm:"type:varchar(255);not null"           json:"-"`
	Avatar   string `gorm:"type:varchar(500)"                    json:"avatar,omitempty"`
	IsActive bool   `gorm:"default:true"                         json:"is_active"`
	Roles    []Role `gorm:"many2many:user_roles"                 json:"roles,omitempty"`
}

type Role struct {
	Base
	Name        string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions"             json:"permissions,omitempty"`
}

type Permission struct {
	Base
	Name string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
}
