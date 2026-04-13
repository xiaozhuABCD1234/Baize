package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique;size:50;not null"` // 用户名
	Email    string `gorm:"unique;not null"`         // 用户邮箱
	Password string `gorm:"not null"`                // 存储哈希后的密码
	Phone    string `gorm:"size:20"`                 // 手机号
	Role     string `gorm:"size:20;default:user"`    // 角色: admin / user
}

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
