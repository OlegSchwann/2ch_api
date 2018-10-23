package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"
)

func (e *Environment)PostsCreate(ctx *fasthttp.RequestCtx) {
	// slug_or_id := ctx.UserValue("slug_or_id").(string)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}

