package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"strings"

	"github.com/OlegSchwann/2ch_api/types"
)

// Получение информации о ветке обсуждения по его имени.
func (e *Environment) PostGetDetails(ctx *fasthttp.RequestCtx) {
	postStringId := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(postStringId) // Идентификатор сообщения.
	if err != nil {
		response, _ := types.Error{
			Message: "post identificator '" + postStringId + "' must be integer",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}
	related := string(ctx.URI().QueryArgs().Peek("related"))
	needUser := strings.Contains(related, "user")
	needForum := strings.Contains(related, "forum")
	needThread := strings.Contains(related, "thread")

	postDetails := types.PostFull{}
	// информация самого поста
	post, err := e.ConnPool.PostGetDetailsSelectPost(postId)
	postDetails.Post = &post
	if err != nil {
	    accessorError := err.(*accessor.Error)
	    if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Can't find Post with id '" + postStringId + "'.",
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
	}
	if needUser {
		// Связанный user.
		user, err := e.ConnPool.UserGetProfile(postDetails.Post.Author)
		postDetails.Author = &user
		if err != nil {
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	if needThread {
		// Связанный thread.
		thread, err := e.ConnPool.ThreadGetDetailsById(postDetails.Post.ThreadId)
		postDetails.Thread = &thread
		if err != nil {
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	if needForum {
		// Связанный forum.
		forum, err := e.ConnPool.ForumGetDetails(postDetails.Post.Forum)
		postDetails.Forum = &forum
		if err != nil {
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	response, _ := postDetails.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
