package http

import (
	"encoding/json"
	"net/http"
	"park_db_course/internal/models"
	"park_db_course/internal/repository"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

type UserHandlersI interface {
	Create(ctx *fasthttp.RequestCtx)
	GetByNickname(ctx *fasthttp.RequestCtx)
	Update(ctx *fasthttp.RequestCtx)
}

type userH struct {
	userRepo repository.UserRepoI
}

func NewUserH(u repository.UserRepoI) UserHandlersI {
	return &userH{
		userRepo: u,
	}
}

func (h *userH) Create(ctx *fasthttp.RequestCtx) {
	var user models.User

	err := easyjson.Unmarshal(ctx.PostBody(), &user)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}
	user.Nickname = ctx.UserValue("nickname").(string)

	// TODO OPTIMIZE
	//h.userRepo.GetByEmailOrNick(user.Email, user.Nickname)
	user1, err1 := h.userRepo.GetByNickname(user.Nickname)
	user2, err2 := h.userRepo.GetByEmail(user.Email)

	if err1 == nil || err2 == nil {
		var users []models.User
		if err1 == nil {
			users = append(users, user1)
		}
		if err2 == nil && user1.About != user2.About {
			users = append(users, user2)
		}
		ctx.SetContentType("application/json")
		body, _ := json.Marshal(users)
		ctx.SetStatusCode(http.StatusConflict)
		ctx.SetBody(body)
		return
	}

	_, err = h.userRepo.Create(user)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
	body, _ := easyjson.Marshal(user)
	ctx.SetBody(body)
}

func (h *userH) GetByNickname(ctx *fasthttp.RequestCtx) {
	user, err := h.userRepo.GetByNickname(ctx.UserValue("nickname").(string))

	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find user by nickname:"})
		ctx.SetBody(body)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(user)
	ctx.SetBody(body)
}

func (h *userH) Update(ctx *fasthttp.RequestCtx) {
	newUserData, err := h.userRepo.GetByNickname(ctx.UserValue("nickname").(string))
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find user by nickname:"})
		ctx.SetBody(body)
		return
	}

	err = json.Unmarshal(ctx.PostBody(), &newUserData)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	checkUser, err := h.userRepo.GetByEmail(newUserData.Email)
	if checkUser.Nickname != "" && checkUser.Nickname != newUserData.Nickname {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusConflict)
		body, _ := easyjson.Marshal(models.MessageError{Message: "This email is already registered by user " + checkUser.Nickname})
		ctx.SetBody(body)
		return
	}

	user, err := h.userRepo.Update(newUserData)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	if newUserData.About == "" {
		user.About = newUserData.About
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(user)
	ctx.SetBody(body)
}
