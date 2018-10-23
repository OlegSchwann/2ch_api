package handlers

import (
	"2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
)

func (e *Environment)ThreadCreate(ctx *fasthttp.RequestCtx) {
	// slug := ctx.UserValue("slug").(string)
	requestThread := types.Thread{}
	err := requestThread.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "unrecognized request: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}

	// TODO: logic

	responseThread := types.Thread{}
	response, err := responseThread.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}

