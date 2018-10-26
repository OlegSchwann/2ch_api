package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ThreadCreate(ctx *fasthttp.RequestCtx) {
	requestThread := types.Thread{}
	err := requestThread.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "unrecognized request: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}
	requestThread.Forum = ctx.UserValue("slug").(string)

	responseThread, err := e.ConnPool.ThreadCreate(requestThread)
	if err != nil {
		accessorError := err.(*accessor.Error)
		switch accessorError.Code {
		case http.StatusNotFound:
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		case http.StatusConflict:
			response, _ := responseThread.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusConflict)
			return
		}
	}
	response, err := responseThread.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
}
