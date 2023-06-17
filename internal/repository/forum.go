package repository

import (
	"fmt"
	"park_db_course/internal/models"

	"github.com/jackc/pgx"
)

type ForumRepoI interface {
	Create(new models.ForumReq) (models.Forum, error)
	GetBySlug(slug string) (forum models.Forum, err error)
	GetThreads(slug, since string, limit int, desc bool) ([]models.Thread, error)
	GetUsers(forum models.Forum, since string, limit int, desc bool) ([]models.User, error)
}

var (
	createForumQ     = `INSERT INTO forum (title, "user", slug) values ($1, $2, $3) RETURNING title, "user", slug, posts, threads;`
	getForumBySlugQ  = `SELECT id, title, "user", slug, posts, threads FROM forum WHERE slug = $1;`
	getForumThreadsQ = `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE forum = $1`
	getForumUsersQ   = `SELECT nickname, about, email, fullname FROM "user" WHERE id IN (SELECT "user" FROM forum_user WHERE forum = $1)`
)

type forumRepo struct {
	db *pgx.ConnPool
}

func NewForumRepo(d *pgx.ConnPool) ForumRepoI {
	return &forumRepo{db: d}
}

func (r *forumRepo) Create(new models.ForumReq) (forum models.Forum, err error) {
	err = r.db.QueryRow(createForumQ, new.Title, new.User, new.Slug).Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
	return
}

func (r *forumRepo) GetBySlug(slug string) (forum models.Forum, err error) {
	err = r.db.QueryRow(getForumBySlugQ, slug).Scan(&forum.Id, &forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)
	return
}

func (r *forumRepo) GetThreads(slug, since string, limit int, desc bool) ([]models.Thread, error) {
	editQuery := getForumThreadsQ
	if since != "" {
		if desc {
			editQuery += fmt.Sprintf(` AND created <= '%s'`, since)
		} else {
			editQuery += fmt.Sprintf(` AND created >= '%s'`, since)
		}
	}
	editQuery += fmt.Sprintf(` ORDER BY created`)
	if desc {
		editQuery += fmt.Sprintf(` DESC`)
	}
	if limit != 0 {
		editQuery += fmt.Sprintf(` LIMIT %d;`, limit)
	}

	rows, err := r.db.Query(editQuery, slug)
	if err != nil {
		return []models.Thread{}, err
	}
	defer rows.Close()

	threads := make([]models.Thread, 0)
	for rows.Next() {
		var t models.Thread
		err = rows.Scan(
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
			return []models.Thread{}, err
		}

		threads = append(threads, t)
	}
	return threads, nil
}

func (r *forumRepo) GetUsers(forum models.Forum, since string, limit int, desc bool) ([]models.User, error) {
	editQuery := getForumUsersQ
	if since != "" {
		if desc {
			editQuery += fmt.Sprintf(` AND nickname < '%s'`, since)
		} else {
			editQuery += fmt.Sprintf(` AND nickname > '%s'`, since)
		}
	}
	editQuery += fmt.Sprintf(` ORDER BY "nickname"`)
	if desc {
		editQuery += fmt.Sprintf(` DESC`)
	}

	if limit > 0 {
		editQuery += fmt.Sprintf(` LIMIT %d;`, limit)
	}

	rows, err := r.db.Query(editQuery, forum.Id)
	if err != nil {
		return []models.User{}, err
	}
	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.Nickname,
			&u.About,
			&u.Email,
			&u.Fullname,
		)
		if err != nil {
			return []models.User{}, err
		}

		users = append(users, u)
	}
	return users, nil
}
