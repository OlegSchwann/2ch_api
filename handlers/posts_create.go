package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"time"
)

func (e *Environment)PostsCreate(ctx *fasthttp.RequestCtx) {
	posts := types.Posts{}
	err := posts.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "unable to parse json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
		return
	}

	if len(posts) == 0 {
		response := []byte("[]")
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusCreated)
		return
	}

	threadslug := ""
	slugOrId := ctx.UserValue("slug_or_id").(string)
	if id, err := strconv.Atoi(slugOrId); err == nil {
		for i := range posts{
			posts[i].Thread = uint(id)
		}
	} else {
		threadslug = slugOrId
	}

	timeNow := time.Now()
	for i := range posts{
		posts[i].Created = timeNow
	}

	responsePosts, err := e.ConnPool.PostsCreate(posts, threadslug)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Not found: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		if accessorError.Code == http.StatusInternalServerError {
			response, _ := types.Error{
				Message: "Server error: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	response, _ := responsePosts.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
	return
}
