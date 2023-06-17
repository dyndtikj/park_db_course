package models

//go:generate easyjson -snake_case -all

type MessageError struct {
	Message string
}
