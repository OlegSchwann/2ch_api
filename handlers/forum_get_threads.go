package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ForumGetThreads(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	if true {
		responseThread := types.Thread{}
		response, _ := responseThread.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusOK)
	} else {
		response, _ := types.Error{
			Message: "Thread '" + slug + "' not found.",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
	}
}
