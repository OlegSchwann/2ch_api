package handlers

import (
	"2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
)

func (e *Environment)ThreadGetDetails(ctx *fasthttp.RequestCtx) {
	slugOrId := ctx.UserValue("slug_or_id").(string)

	// TODO: logic

	if true {
		responseForum := types.Forum{}
		response, _ := responseForum.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusOK)
	} else {
		response, _ := types.Error{
			Message: "Thread '" + slugOrId + "' not found.",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
	}
}

