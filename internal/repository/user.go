package repository

import (
	"park_db_course/internal/models"

	"github.com/jackc/pgx"
)

type UserRepoI interface {
	Create(newUser models.User) (models.User, error)
	GetByNickname(nickname string) (user models.User, err error)
	GetByEmail(email string) (user models.User, err error)
	GetByEmailOrNick(email, nickname string) (users []*models.User, err error)
	Update(user models.User) (NewUser models.User, err error)
}

var (
	createUserQ           = `INSERT INTO "user" (nickname, fullname, about, email) VALUES ($1, $2, $3, $4) RETURNING id, nickname, fullname, about, email;`
	getUserByNicknameQ    = `SELECT id, nickname, fullname, about, email FROM "user" WHERE nickname = $1;`
	getUserByEmailQ       = `SELECT id, nickname, fullname, about, email  FROM "user" WHERE email = $1;`
	getUserByEmailOrNickQ = `SELECT id, nickname, fullname, about, email FROM "user" WHERE nickname = $1 OR email = $2;`
	updateUserQ           = `UPDATE "user" SET fullname = $2, about = $3, email = $4 WHERE nickname = $1 RETURNING nickname, fullname, about, email;`
)

type userRepo struct {
	db *pgx.ConnPool
}

func NewUserRepo(d *pgx.ConnPool) UserRepoI {
	return &userRepo{db: d}
}

func (r *userRepo) Create(newUser models.User) (user models.User, err error) {
	err = r.db.QueryRow(createUserQ, newUser.Nickname, newUser.Fullname, newUser.About, newUser.Email).Scan(&user.Id, &user.Nickname, &user.Fullname, &user.About, &user.Email)
	return
}

func (r *userRepo) GetByNickname(nickname string) (user models.User, err error) {
	err = r.db.QueryRow(getUserByNicknameQ, nickname).Scan(&user.Id, &user.Nickname, &user.Fullname, &user.About, &user.Email)
	return
}

func (r *userRepo) GetByEmail(email string) (user models.User, err error) {
	err = r.db.QueryRow(getUserByEmailQ, email).Scan(&user.Id, &user.Nickname, &user.Fullname, &user.About, &user.Email)
	return
}

func (r *userRepo) GetByEmailOrNick(email, nickname string) (users []*models.User, err error) {
	rows, err := r.db.Query(getUserByEmailOrNickQ, nickname, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		u := &models.User{}
		err = rows.Scan(&u.Id, &u.Nickname, &u.Fullname, &u.About, &u.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepo) Update(user models.User) (NewUser models.User, err error) {
	err = r.db.QueryRow(updateUserQ, user.Nickname, user.Fullname, user.About, user.Email).Scan(&NewUser.Nickname, &NewUser.Fullname, &NewUser.About, &NewUser.Email)
	return
}
