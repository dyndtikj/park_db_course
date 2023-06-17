package repository

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"park_db_course/internal/models"

	"github.com/jackc/pgx"
)

type ThreadRepoI interface {
	GetBySlugOrId(slug string) (t models.Thread, err error)
	Create(new models.ThreadsReq) (t models.Thread, err error)
	Update(old models.Thread, new models.ThreadUpdateReq) (t models.Thread, err error)
	CheckPost(parent, id int) (err error)
	CreatePosts(thread models.Thread, new models.PostsReq) (response *models.Posts, err error)
	CheckVotes(user, thread int) (vote models.Vote, err error)
	CreateVote(userId int, vote models.VoteRequest, thread models.Thread) (err error)
	UpdateVote(vote models.VoteRequest, voteId int) (id int, err error)
	GetThreadPosts(thread models.Thread, since, sort string, limit int, desc bool) ([]models.Post, error)
}

var (
	createThreadQ      = `INSERT INTO thread (title, author, forum, message, slug, created) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, author, forum, message, votes, slug, created;`
	updateThreadQ      = `UPDATE thread SET title = $1, message = $2 WHERE id = $3 RETURNING id, title, author, forum, message, votes, slug, created;`
	getThreadQ         = `SELECT id, title, author, forum, message, votes, slug, created FROM thread WHERE slug = $1 OR id = $2;`
	checkThreadPostQ   = `SELECT id FROM post WHERE thread = $1 AND id = $2;`
	createThreadPostsQ = `INSERT INTO post (parent, author, message, forum, thread, created) values `
	checkVotesQ        = `SELECT id, "user", thread, voice from vote where "user" = $1 and thread = $2;`
	createVoteQ        = `INSERT INTO vote ("user", thread, voice)  VALUES ($1, $2, $3)  RETURNING "user";`
	updateVoteQ        = `UPDATE vote SET voice = $1 WHERE id = $2 RETURNING id;`
	getThreadPostsQ    = `SELECT id, parent, author, message, is_edited, forum, thread, created FROM post WHERE thread = $1 `
)

type threadRepo struct {
	db *pgx.ConnPool
}

func NewThreadRepo(d *pgx.ConnPool) ThreadRepoI {
	return &threadRepo{db: d}
}

func (r *threadRepo) Create(new models.ThreadsReq) (t models.Thread, err error) {
	if new.Created.String() == "" {
		new.Created = time.Now()
	}

	err = r.db.QueryRow(createThreadQ, new.Title, new.Author, new.Forum, new.Message, new.Slug, new.Created).Scan(
		&t.Id, &t.Title, &t.Author, &t.Forum, &t.Message, &t.Votes, &t.Slug, &t.Created)
	return
}

func (r *threadRepo) Update(oldThread models.Thread, newThread models.ThreadUpdateReq) (t models.Thread, err error) {
	err = r.db.QueryRow(updateThreadQ, newThread.Title, newThread.Message, oldThread.Id).Scan(&t.Id, &t.Title, &t.Author, &t.Forum, &t.Message, &t.Votes, &t.Slug, &t.Created)
	return
}

func (r *threadRepo) GetBySlugOrId(slug string) (t models.Thread, err error) {
	id, _ := strconv.Atoi(slug)
	err = r.db.QueryRow(getThreadQ, slug, id).Scan(&t.Id, &t.Title, &t.Author, &t.Forum, &t.Message, &t.Votes, &t.Slug, &t.Created)
	return
}

func (r *threadRepo) CheckPost(parent, id int) (err error) {
	err = r.db.QueryRow(checkThreadPostQ, id, parent).Scan(&id)
	return
}

func (r *threadRepo) CreatePosts(thread models.Thread, new models.PostsReq) (response *models.Posts, err error) {
	editQuery := createThreadPostsQ

	var postsValues []interface{}
	created := time.Now()

	for i, post := range new.Posts {
		if i != 0 {
			editQuery += ", "
		}
		editQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", 1+i*6, 2+i*6, 3+i*6, 4+i*6, 5+i*6, 6+i*6)
		postsValues = append(postsValues, post.Parent, post.Author, post.Message, thread.Forum, thread.Id, created)
	}

	editQuery += ` RETURNING id, parent, author, message, is_edited, forum, thread, created;`

	rows, err := r.db.Query(editQuery, postsValues...)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	response = &models.Posts{}

	for rows.Next() {
		var p models.Post
		err = rows.Scan(
			&p.Id,
			&p.Parent,
			&p.Author,
			&p.Message,
			&p.IsEdited,
			&p.Forum,
			&p.Thread,
			&p.Created,
		)
		if err != nil {
			return nil, errors.New(err.Error())
		}

		response.Posts = append(response.Posts, p)
	}

	return response, nil
}

func (r *threadRepo) CheckVotes(user, thread int) (vote models.Vote, err error) {
	err = r.db.QueryRow(checkVotesQ, user, thread).Scan(&vote.Id, &vote.User, &vote.Thread, &vote.Voice)
	return
}

func (r *threadRepo) CreateVote(userId int, vote models.VoteRequest, thread models.Thread) (err error) {
	err = r.db.QueryRow(createVoteQ, userId, thread.Id, vote.Voice).Scan(&userId)
	return
}

func (r *threadRepo) UpdateVote(vote models.VoteRequest, voteId int) (id int, err error) {
	err = r.db.QueryRow(updateVoteQ, vote.Voice, voteId).Scan(&id)
	return
}

func (r *threadRepo) GetThreadPosts(thread models.Thread, since, sort string, limit int, desc bool) ([]models.Post, error) {
	posts := make([]models.Post, 0)

	editQuery := getThreadPostsQ

	cmp := ">"
	order := "ASC"
	if desc {
		cmp = "<"
		order = "DESC"
	}

	switch sort {
	case "flat":
		if since != "" {
			editQuery += fmt.Sprintf(` AND id %s %s `, cmp, since)
		}
		editQuery += fmt.Sprintf(` ORDER BY created %s, id %s LIMIT %d `, order, order, limit)
	case "tree":
		if since != "" {
			editQuery += fmt.Sprintf(` AND path %s (SELECT path FROM post WHERE id = %s) `, cmp, since)
		}
		editQuery += fmt.Sprintf(` ORDER BY path[1] %s, path %s limit %d `, order, order, limit)
	case "parent_tree":
		editQuery += ` AND path && (SELECT ARRAY (SELECT id FROM post WHERE thread = $1 and parent = 0 `
		if since != "" {
			editQuery += fmt.Sprintf(` AND path %s (SELECT path[1:1] FROM post WHERE id = %s) `, cmp, since)
		}
		editQuery += fmt.Sprintf(` ORDER BY path[1] %s, path LIMIT %d)) ORDER BY path[1] %s, path `, order, limit, order)
	default:
		return []models.Post{}, errors.New("wrong sort name")
	}

	rows, err := r.db.Query(editQuery, thread.Id)
	if err != nil {
		return []models.Post{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.Id, &p.Parent, &p.Author, &p.Message, &p.IsEdited, &p.Forum, &p.Thread, &p.Created)
		if err != nil {
			return []models.Post{}, err
		}

		posts = append(posts, p)
	}

	return posts, nil
}
