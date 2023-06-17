package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"park_db_course/internal/models"
	"park_db_course/internal/repository"
	"strconv"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

type ForumHandlersI interface {
	Create(ctx *fasthttp.RequestCtx)
	Details(ctx *fasthttp.RequestCtx)
	CreateThread(ctx *fasthttp.RequestCtx)
	ForumThreads(ctx *fasthttp.RequestCtx)
	ForumUsers(ctx *fasthttp.RequestCtx)
}

type forumH struct {
	forumRepo  repository.ForumRepoI
	userRepo   repository.UserRepoI
	threadRepo repository.ThreadRepoI
}

func NewForumH(f repository.ForumRepoI, u repository.UserRepoI, t repository.ThreadRepoI) ForumHandlersI {
	return &forumH{
		forumRepo:  f,
		userRepo:   u,
		threadRepo: t,
	}
}

func (h *forumH) Create(ctx *fasthttp.RequestCtx) {
	forum := models.ForumReq{}
	err := easyjson.Unmarshal(ctx.PostBody(), &forum)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	checkForum, err := h.forumRepo.GetBySlug(forum.Slug)
	if err == nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusConflict)
		body, _ := easyjson.Marshal(checkForum)
		ctx.SetBody(body)
		return
	}

	checkUser, err := h.userRepo.GetByNickname(forum.User)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{})
		ctx.SetBody(body)
		return
	}

	forum.User = checkUser.Nickname

	newForum, err := h.forumRepo.Create(forum)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
	body, err := easyjson.Marshal(newForum)
	ctx.SetBody(body)
}

func (h *forumH) Details(ctx *fasthttp.RequestCtx) {
	checkForum, err := h.forumRepo.GetBySlug(ctx.UserValue("slug").(string))
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{})
		ctx.SetBody(body)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(checkForum)
	ctx.SetBody(body)
}

func (h *forumH) CreateThread(ctx *fasthttp.RequestCtx) {
	checkForum, err := h.forumRepo.GetBySlug(ctx.UserValue("slug").(string))
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{})
		ctx.SetBody(body)
		return
	}

	thread := models.ThreadsReq{}
	err = easyjson.Unmarshal(ctx.PostBody(), &thread)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	if thread.Slug != "" {
		checkThread, err := h.threadRepo.GetBySlugOrId(thread.Slug)
		if err == nil {
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusConflict)
			body, _ := easyjson.Marshal(checkThread)
			ctx.SetBody(body)
			return
		}
	}

	thread.Forum = checkForum.Slug

	checkAuthor, err := h.userRepo.GetByNickname(thread.Author)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{})
		ctx.SetBody(body)
		return
	}
	thread.Author = checkAuthor.Nickname

	newThread, err := h.threadRepo.Create(thread)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusInternalServerError)
		fmt.Println("WHAT", err)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
	body, _ := easyjson.Marshal(newThread)
	ctx.SetBody(body)
}

func (h *forumH) ForumThreads(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	_, err := h.forumRepo.GetBySlug(slug)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{})
		ctx.SetBody(body)
		return
	}

	limit := 0
	if limitVal := string(ctx.FormValue("limit")); limitVal != "" {
		limit, err = strconv.Atoi(limitVal)
		if err != nil {
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusBadRequest)
			body, _ := easyjson.Marshal(models.MessageError{Message: "wrong limit format"})
			ctx.SetBody(body)
		}
	} else {
		limit = 100
	}

	since := string(ctx.FormValue("since"))

	desc := false
	if string(ctx.FormValue("desc")) == "true" {
		desc = true
	}

	threads, err := h.forumRepo.GetThreads(slug, since, limit, desc)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := json.Marshal(threads)
	ctx.SetBody(body)
}

func (h *forumH) ForumUsers(ctx *fasthttp.RequestCtx) {
	slug, ok := ctx.UserValue("slug").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug format"})
		ctx.SetBody(body)
	}
	forum, err := h.forumRepo.GetBySlug(slug)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find forum by slug: " + slug})
		ctx.SetBody(body)
		return
	}

	limit := 0
	if limitVal := string(ctx.FormValue("limit")); limitVal != "" {
		limit, err = strconv.Atoi(limitVal)
		if err != nil {
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusBadRequest)
			body, _ := easyjson.Marshal(models.MessageError{Message: "wrong limit format"})
			ctx.SetBody(body)
		}
	} else {
		limit = 100
	}

	since := string(ctx.FormValue("since"))

	desc := false
	if string(ctx.FormValue("desc")) == "true" {
		desc = true
	}

	users, err := h.forumRepo.GetUsers(forum, since, limit, desc)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := json.Marshal(users)
	ctx.SetBody(body)
}
