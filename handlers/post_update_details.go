package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"
)

func (e *Environment)PostUpdateDetails(ctx *fasthttp.RequestCtx) {
	// id := ctx.UserValue("id").(string)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}

