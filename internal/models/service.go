package models

//go:generate easyjson -snake_case -all

type Status struct {
	User   int
	Forum  int
	Thread int
	Post   int
}
