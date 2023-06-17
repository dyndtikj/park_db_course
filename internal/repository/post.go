package repository

import (
	"park_db_course/internal/models"

	"github.com/jackc/pgx"
)

type PostRepoI interface {
	Get(id int, related []string) (postInfo models.PostFull, err error)
	Update(id int, new models.PostUpdateReq) (p models.Post, err error)
}

var (
	getPostQ       = `SELECT id, parent, author, message, is_edited, forum, thread, created FROM post WHERE id = $1;`
	getPostUserQ   = `SELECT nickname, fullname, about, email FROM "user" WHERE nickname = $1;`
	getPostForumQ  = `SELECT title, "user", slug, posts, threads FROM forum WHERE slug = $1;`
	getPostThreadQ = `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE id = $1;`
	updatePostQ    = `UPDATE post SET message = $1, is_edited = TRUE WHERE id = $2 RETURNING id, parent, author, message, is_edited, forum, thread, created;`
)

type postRepo struct {
	db *pgx.ConnPool
}

func NewPostRepo(d *pgx.ConnPool) PostRepoI {
	return &postRepo{db: d}
}

func (r *postRepo) Get(id int, related []string) (postInfo models.PostFull, err error) {
	var post models.Post
	err = r.db.QueryRow(getPostQ, id).Scan(
		&post.Id,
		&post.Parent,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&post.Thread,
		&post.Created,
	)
	if err != nil {
		return
	}

	postInfo.Post = &post

	if len(related) != 0 {
		for _, q := range related {
			switch q {
			case "user":
				var u models.User
				err = r.db.QueryRow(getPostUserQ, post.Author).Scan(
					&u.Nickname,
					&u.Fullname,
					&u.About,
					&u.Email,
				)
				if err != nil {
					return
				}
				postInfo.Author = &u
			case "forum":
				var f models.Forum
				err = r.db.QueryRow(getPostForumQ, post.Forum).Scan(
					&f.Title,
					&f.User,
					&f.Slug,
					&f.Posts,
					&f.Threads,
				)
				if err != nil {
					return
				}
				postInfo.Forum = &f
			case "thread":
				var t models.Thread
				err = r.db.QueryRow(getPostThreadQ, post.Thread).Scan(
					&t.Id,
					&t.Title,
					&t.Author,
					&t.Forum,
					&t.Message,
					&t.Votes,
					&t.Slug,
					&t.Created,
				)
				if err != nil {
					return
				}
				postInfo.Thread = &t
			default:
				break
			}
		}
	}

	return
}

func (r *postRepo) Update(id int, new models.PostUpdateReq) (p models.Post, err error) {
	err = r.db.QueryRow(updatePostQ, new.Message, id).Scan(
		&p.Id,
		&p.Parent,
		&p.Author,
		&p.Message,
		&p.IsEdited,
		&p.Forum,
		&p.Thread,
		&p.Created,
	)
	return
}
