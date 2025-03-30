package models

import (
	"github.com/Brain-Wave-Ecosystem/go-common/pkg/helpers"
	users "github.com/Brain-Wave-Ecosystem/users-service/gen/users"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"unicode"
)

func (u *User) ToGRPC() *users.User {
	user := &users.User{
		Id:         u.ID,
		Email:      u.Email,
		AvatarUrl:  u.AvatarURL,
		FullName:   u.FullName,
		Slug:       u.Slug,
		Bio:        u.Bio,
		Role:       u.Role,
		IsVerified: u.IsVerified,
		CreatedAt:  timestamppb.New(u.CreatedAt),
	}

	if u.LastLoginAt != nil {
		user.LastLoginAt = timestamppb.New(*u.LastLoginAt)
	}

	if u.UpdatedAt != nil {
		user.UpdatedAt = timestamppb.New(*u.UpdatedAt)
	}

	return user
}

func ToUserWithPassword(r *users.CreateUserRequest) *UserWithPassword {
	return &UserWithPassword{
		User: &User{
			Email:    r.Email,
			FullName: r.FullName,
		},
		UserPassword: &UserPassword{
			PasswordHash: r.Password,
		},
	}
}

func (u *UserWithPassword) PrepareUser() *UserWithPassword {
	u.FullName = prepareFullName(u.FullName)
	u.Slug = helpers.GenerateSlug(u.FullName)
	return u
}

func (u *UpdateUser) PrepareUser() *UpdateUser {
	if u.FullName != nil {
		newName := prepareFullName(*u.FullName)
		u.FullName = &newName

		u.Slug = helpers.GenerateSlug(newName)
	}
	return u
}

func prepareFullName(fullName string) string {
	names := strings.Fields(fullName)

	if len(names) > 0 {
		for i, name := range names {
			runes := []rune(name)
			runes[0] = unicode.ToUpper(runes[0])
			names[i] = string(runes)
		}
	}

	return strings.Join(names, " ")
}
