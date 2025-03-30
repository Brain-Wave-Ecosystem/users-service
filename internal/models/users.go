package models

import "time"

type User struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	AvatarURL   *string    `json:"avatarUrl,omitempty"`
	FullName    string     `json:"fullName,omitempty"`
	Slug        string     `json:"slug"`
	Bio         *string    `json:"bio,omitempty"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	Role        string     `json:"role"`
	IsVerified  *bool      `json:"isVerified,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type UserWithPassword struct {
	*User
	*UserPassword
}

type UserPassword struct {
	PasswordHash string     `json:"passwordHash"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

type UserPasswordHistory struct {
	PasswordHash string    `json:"passwordHash" db:"pass_hash"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type UpdateUser struct {
	Slug      string  `json:"slug"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	FullName  *string `json:"fullName,omitempty"`
	Bio       *string `json:"bio,omitempty"`
}
