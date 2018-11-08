package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

// Получение списка пользователей, у которых есть пост или ветка обсуждения в данном форуме.
// Пользователи выводятся отсортированные по nickname в порядке возрастания.
// Порядок сотрировки должен соответсвовать побайтовому сравнение в нижнем регистре.
func (e *Environment) ForumGetUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	queryArgs := ctx.Request.URI().QueryArgs()
	limit, err := queryArgs.GetUint("limit")
	if err != nil {
		limit = 100
	}
	since := string(queryArgs.Peek("since"))
	desc := queryArgs.GetBool("desc")

		exits, err := e.ConnPool.ForumGetThreadsCheckForumExist(slug)
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	if !exits {
		response, _ := types.Error{
			Message: "Can't find forum with slug '" + slug + "'.",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
		return
	}

	var users types.Users
	if since != "" {
		users, err = e.ConnPool.ForumGetUsersSince(slug, limit, since, desc)
	} else {
		users, err = e.ConnPool.ForumGetUsers(slug, limit, desc)
	}
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	response := []byte("[]")
	if len(users) != 0 {
		response, _ = users.MarshalJSON()
	}
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
