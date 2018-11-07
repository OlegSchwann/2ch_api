package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ThreadGetPosts(ctx *fasthttp.RequestCtx) {
	slugOrId := ctx.UserValue("slug_or_id").(string) // Идентификатор ветки обсуждения.
	// находим id по slug или отдаём 404.
	// в случае, если обращаются с id в url, всё равно делаем поход в базу, проверяя существование thread.
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		threadId, err = e.ConnPool.ThreadVoteGetThreadIdBySlug(slugOrId)
		if err != nil {
			accessorError := err.(*accessor.Error)
			if accessorError.Code == http.StatusNotFound {
				response, _ := types.Error{
					Message: "not found thread '" + slugOrId + "'",
				}.MarshalJSON()
				ctx.Write(response)
				ctx.Response.Header.SetStatusCode(http.StatusNotFound)
				return
			}
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	} else {
		threadExists, err := e.ConnPool.ThreadGetPostsCheckIfThreadExists(threadId)
		if err != nil {
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
		if !threadExists {
			response, _ := types.Error{
				Message: "not found thread '" + slugOrId + "'",
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
	}

	args := ctx.URI().QueryArgs()
	limit, err := args.GetUint("limit") // Максимальное кол-во возвращаемых записей.
	if err != nil {
		limit = 100
	}

	since, err := args.GetUint("since") // Идентификатор поста, после которого будут выводиться записи
	if err != nil { //                         пост с данным идентификатором в результат не попадает).
		since = -1
	}

	desc := args.GetBool("desc")

	posts := types.Posts{}
	switch string(args.Peek("sort")) {
	case "":
		fallthrough // По умолчанию: flat
	case "flat": // по дате, комментарии выводятся простым списком в порядке создания;
		if since != -1 {
			posts, err = e.ConnPool.ThreadGetPostsFlatSince(threadId, limit, since, desc)
		} else {
			posts, err = e.ConnPool.ThreadGetPostsFlatSort(threadId, limit, desc)
		}
	case "tree": // древовидный, комментарии выводятся отсортированные в дереве по N штук;
		if since != -1 {
			posts, err = e.ConnPool.ThreadGetPostsTreeSince(threadId, limit, since, desc)
		} else {
			posts, err = e.ConnPool.ThreadGetPostsTree(threadId, limit, desc)
		}
	case "parent_tree": // древовидные с пагинацией по родительским (parent_tree), на странице N родительских комментов и все комментарии прикрепленные к ним, в древвидном отображение. Подробности: https://park.mail.ru/blog/topic/view/1191/
		if since != -1 {
			if desc {
				posts, err = e.ConnPool.ThreadGetPostsParentTreeSinceSortDesc(threadId, limit, since)
			} else {
				posts, err = e.ConnPool.ThreadGetPostsParentTreeSinceSortAsc(threadId, limit, since)
			}
		} else {
			if desc {
				posts, err = e.ConnPool.ThreadGetPostsParentTreeSortDesc(threadId, limit)
			} else {
				posts, err = e.ConnPool.ThreadGetPostsParentTreeSortAsc(threadId, limit)
			}
		}
	}
	response, err := posts.MarshalJSON()
	if len(posts) == 0 {
		response = []byte("[]")
	}
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}
