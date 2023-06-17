package models

//go:generate easyjson -snake_case -all

type User struct {
	Id       int `json:"-"`
	Nickname string
	Fullname string
	About    string
	Email    string
}

type Users struct {
	Users []User
}

type UserUpdate struct {
	Fullname string
	About    string
	Email    string
}
