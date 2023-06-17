package http

import (
	"net/http"
	"strconv"
	"strings"

	"park_db_course/internal/models"
	"park_db_course/internal/repository"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

type PostHandlersI interface {
	GetDetails(ctx *fasthttp.RequestCtx)
	UpdateDetails(ctx *fasthttp.RequestCtx)
}

type postH struct {
	postRepo repository.PostRepoI
}

func NewPostH(p repository.PostRepoI) PostHandlersI {
	return &postH{
		postRepo: p,
	}
}

func (h *postH) GetDetails(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	related := strings.Split(string(ctx.FormValue("related")), ",")
	post, err := h.postRepo.Get(id, related)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find post with id: " + strconv.Itoa(id)})
		ctx.SetBody(body)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(post)
	ctx.SetBody(body)
}

func (h *postH) UpdateDetails(ctx *fasthttp.RequestCtx) {
	id, err := strconv.Atoi(ctx.UserValue("id").(string))
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	var newPost models.PostUpdateReq
	err = easyjson.Unmarshal(ctx.PostBody(), &newPost)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	var related []string
	postInfo, err := h.postRepo.Get(id, related)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find post with id: " + strconv.Itoa(id)})
		ctx.SetBody(body)
		return
	}

	oldPost := postInfo.Post
	if newPost.Message == "" || oldPost.Message == newPost.Message {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusOK)
		body, _ := easyjson.Marshal(oldPost)
		ctx.SetBody(body)
		return
	}

	post, err := h.postRepo.Update(id, newPost)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(post)
	ctx.SetBody(body)
	return
}
