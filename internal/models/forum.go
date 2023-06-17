package models

//go:generate easyjson -snake_case -all

type ForumReq struct {
	Title string
	User  string
	Slug  string
}

type Forum struct {
	Id      int64 `json:"-"`
	Title   string
	User    string
	Slug    string
	Posts   int
	Threads int
}
