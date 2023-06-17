package models

import "time"

//go:generate easyjson -snake_case -all

type ThreadsReq struct {
	Title   string
	Author  string
	Message string
	Created time.Time `json:",omitempty"`
	Forum   string
	Slug    string
}

type ThreadUpdateReq struct {
	Title   string
	Message string
}

type Thread struct {
	Id      int
	Title   string
	Author  string
	Forum   string
	Message string
	Votes   int
	Slug    string
	Created time.Time
}
