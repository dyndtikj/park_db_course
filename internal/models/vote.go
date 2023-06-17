package models

//go:generate easyjson  -snake_case -all

type VoteRequest struct {
	Nickname string
	Voice    int
}

type Vote struct {
	Id     int `json:"-"`
	User   int
	Thread int
	Voice  int
}
