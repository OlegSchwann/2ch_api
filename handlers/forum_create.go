package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ForumCreate(ctx *fasthttp.RequestCtx) {
	if ctx.UserValue("slug") != "create" {
		ctx.NotFound()
		return
	}
	requestForum := types.Forum{}
	err := requestForum.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "unrecognized request: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}

	responseForum, err := e.ConnPool.ForumCreate(requestForum)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusConflict {
			responseForum, err := e.ConnPool.ForumCreateOnConflict(requestForum.Slug)
			if err != nil {
				response, _ := types.Error{
					Message: "Internal server error: " + err.Error(),
				}.MarshalJSON()
				ctx.Write(response)
				ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
				return
			}
			response, err := responseForum.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusConflict)
			return
		}
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "user '" + requestForum.User + "' not found: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
	}
	response, err := responseForum.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
	return
}
