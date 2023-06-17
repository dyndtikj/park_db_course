package repository

import (
	"park_db_course/internal/models"

	"github.com/jackc/pgx"
)

type ServiceRepoI interface {
	Status() (models.Status, error)
	Clear() error
}

var (
	getDBInfoQ = `SELECT (SELECT count(*) from forum), (SELECT count(*) from post), (SELECT count(*) from thread), (SELECT count(*) from "user");`
	deleteDBQ  = `TRUNCATE "user", forum, thread, post, vote, forum_user CASCADE;`
)

type serviceRepo struct {
	db *pgx.ConnPool
}

func NewServiceRepo(d *pgx.ConnPool) ServiceRepoI {
	return &serviceRepo{db: d}
}

func (r *serviceRepo) Status() (models.Status, error) {
	var status models.Status
	err := r.db.QueryRow(getDBInfoQ).Scan(
		&status.Forum,
		&status.Post,
		&status.Thread,
		&status.User,
	)
	return status, err
}

func (r *serviceRepo) Clear() error {
	err := r.db.QueryRow(deleteDBQ).Scan()
	return err
}
