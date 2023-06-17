package http

import (
	"encoding/json"
	"net/http"
	"park_db_course/internal/models"
	"park_db_course/internal/repository"
	"strconv"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

type ThreadHandlersI interface {
	CreatePost(ctx *fasthttp.RequestCtx)
	CreateVote(ctx *fasthttp.RequestCtx)
	Details(ctx *fasthttp.RequestCtx)
	ThreadPost(ctx *fasthttp.RequestCtx)
	Update(ctx *fasthttp.RequestCtx)
}

type threadH struct {
	threadRepo repository.ThreadRepoI
	userRepo   repository.UserRepoI
}

func NewThreadH(t repository.ThreadRepoI, u repository.UserRepoI) ThreadHandlersI {
	return &threadH{threadRepo: t, userRepo: u}
}

func (h *threadH) CreatePost(ctx *fasthttp.RequestCtx) {
	slugOrId, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug_or_id format"})
		ctx.SetBody(body)
	}

	thread, err := h.threadRepo.GetBySlugOrId(slugOrId)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find post thread by id: " + slugOrId})
		ctx.SetBody(body)
		return
	}

	var posts models.PostsReq
	err = json.Unmarshal(ctx.PostBody(), &posts.Posts)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)

		return
	}

	if len(posts.Posts) == 0 {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusCreated)
		body, _ := json.Marshal(posts.Posts)
		ctx.SetBody(body)
		return
	}

	for _, item := range posts.Posts {
		_, err := h.userRepo.GetByNickname(item.Author)
		if err != nil {
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusNotFound)
			body, _ := easyjson.Marshal(models.MessageError{Message: "user by nickname:"})
			ctx.SetBody(body)
			return
		}

		if item.Parent != 0 {
			err = h.threadRepo.CheckPost(item.Parent, thread.Id)
			if err != nil {
				ctx.SetContentType("application/json")
				ctx.SetStatusCode(http.StatusConflict)
				body, _ := easyjson.Marshal(models.MessageError{Message: "post exists"})
				ctx.SetBody(body)
				return
			}
		}
	}

	response, err := h.threadRepo.CreatePosts(thread, posts)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusConflict)
		body, _ := easyjson.Marshal(models.MessageError{Message: err.Error()})
		ctx.SetBody(body)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusCreated)
	body, _ := json.Marshal(response.Posts)
	ctx.SetBody(body)
}

func (h *threadH) CreateVote(ctx *fasthttp.RequestCtx) {
	slugOrId, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug_or_id format"})
		ctx.SetBody(body)
	}

	thread, err := h.threadRepo.GetBySlugOrId(slugOrId)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find thread by slug:  " + slugOrId})
		ctx.SetBody(body)
		return
	}

	var vote models.VoteRequest
	err = easyjson.Unmarshal(ctx.PostBody(), &vote)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	checkUser, err := h.userRepo.GetByNickname(vote.Nickname)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "user by nickname:"})
		ctx.SetBody(body)
		return
	}

	vote1, err := h.threadRepo.CheckVotes(checkUser.Id, thread.Id)
	if err == nil && vote.Voice == vote1.Voice {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusOK)
		body, _ := easyjson.Marshal(thread)
		ctx.SetBody(body)
		return
	}
	if err != nil {
		err = h.threadRepo.CreateVote(checkUser.Id, vote, thread)
		if err != nil {
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}

		thread.Votes += vote.Voice
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusOK)
		body, _ := easyjson.Marshal(thread)
		ctx.SetBody(body)
	} else {
		_, err = h.threadRepo.UpdateVote(vote, vote1.Id)
		if err == nil {
			thread.Votes += 2 * vote.Voice
			ctx.SetContentType("application/json")
			ctx.SetStatusCode(http.StatusOK)
			body, _ := easyjson.Marshal(thread)
			ctx.SetBody(body)
			return
		}
	}
}

func (h *threadH) Details(ctx *fasthttp.RequestCtx) {
	slugOrId, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug_or_id format"})
		ctx.SetBody(body)
	}

	thread, err := h.threadRepo.GetBySlugOrId(slugOrId)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find thread by slug: " + slugOrId})
		ctx.SetBody(body)
		return
	}

	ctx.SetContentType("application/json")
	body, _ := easyjson.Marshal(thread)
	ctx.SetStatusCode(http.StatusOK)
	ctx.SetBody(body)
}

func (h *threadH) ThreadPost(ctx *fasthttp.RequestCtx) {
	slugOrId, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug_or_id format"})
		ctx.SetBody(body)
	}

	thread, err := h.threadRepo.GetBySlugOrId(slugOrId)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find thread by slug: " + slugOrId})
		ctx.SetBody(body)
		return
	}

	sort := string(ctx.FormValue("sort"))
	if sort == "" {
		sort = "flat"
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

	posts, err := h.threadRepo.GetThreadPosts(thread, since, sort, limit, desc)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := json.Marshal(posts)
	ctx.SetBody(body)
}

func (h *threadH) Update(ctx *fasthttp.RequestCtx) {
	slugOrId, ok := ctx.UserValue("slug_or_id").(string)
	if !ok {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		body, _ := easyjson.Marshal(models.MessageError{Message: "wrong slug_or_id format"})
		ctx.SetBody(body)
	}

	thread, err := h.threadRepo.GetBySlugOrId(slugOrId)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		body, _ := easyjson.Marshal(models.MessageError{Message: "Can't find thread by slug: " + slugOrId})
		ctx.SetBody(body)
		return
	}

	var updateThread models.ThreadUpdateReq
	err = easyjson.Unmarshal(ctx.PostBody(), &updateThread)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusBadRequest)
		return
	}

	if updateThread.Title == "" && updateThread.Message == "" {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusOK)
		body, _ := easyjson.Marshal(thread)
		ctx.SetBody(body)
		return
	}
	if updateThread.Title == "" {
		updateThread.Title = thread.Title
	}
	if updateThread.Message == "" {
		updateThread.Message = thread.Message
	}

	thread, err = h.threadRepo.Update(thread, updateThread)
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(thread)
	ctx.SetBody(body)
}
