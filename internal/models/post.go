package models

import "time"

//go:generate easyjson -snake_case -all

type PostReq struct {
	Parent  int
	Author  string
	Message string
}

type PostUpdateReq struct {
	Message string
}

type PostsReq struct {
	Posts []PostReq
}

type Post struct {
	Id       int64
	Parent   int64
	Author   string
	Message  string
	IsEdited bool `json:"isEdited"`
	Forum    string
	Thread   int32
	Created  time.Time
	Path     int64 `json:"-"`
}

type Posts struct {
	Posts []Post
}

type PostFull struct {
	Post   *Post
	Author *User
	Thread *Thread
	Forum  *Forum
}
