package model

import "time"

type Role struct {
	ID        int8      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	Name      string    `gorm:"unique" json:"name"`
}

type User struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Name      string
	Email     string `gorm:"unique"`
	Password  string `json:"password,omitempty"`
	RoleID    int
	Role      Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
