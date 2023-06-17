package http

import (
	"net/http"

	"park_db_course/internal/repository"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

type ServiceHandlersI interface {
	Status(ctx *fasthttp.RequestCtx)
	Clear(ctx *fasthttp.RequestCtx)
}

type serviceH struct {
	serviceRepo repository.ServiceRepoI
}

func NewServiceH(s repository.ServiceRepoI) ServiceHandlersI {
	return &serviceH{serviceRepo: s}
}

func (h *serviceH) Status(ctx *fasthttp.RequestCtx) {
	status, err := h.serviceRepo.Status()
	if err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusOK)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
	body, _ := easyjson.Marshal(status)
	ctx.SetBody(body)
}

func (h *serviceH) Clear(ctx *fasthttp.RequestCtx) {
	if err := h.serviceRepo.Clear(); err != nil {
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(http.StatusInternalServerError)
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(http.StatusOK)
}
