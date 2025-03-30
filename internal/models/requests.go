package models

type GetUserByIDRequest struct {
	ID int `json:"-" param:"userId"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
