package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

func (e *Environment) ForumGetThreads(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	var limit int
	var since time.Time
	var desc bool
	{
		var err error
		queryArgs := ctx.Request.URI().QueryArgs()
		limit, err = queryArgs.GetUint("limit")
		if err != nil || limit > 100 {
			limit = 100
		}
		desc = queryArgs.GetBool("desc")
		sinceBytes := queryArgs.Peek("since")
		since, err = time.Parse(time.RFC3339, string(sinceBytes))
		if err != nil {
			if desc {
				since = time.Unix(64060588800, 0)
			} else {
				since = time.Unix(0, 0)
			}
		}
	}
	threads, err := e.ConnPool.ForumGetThreads(slug, limit, since, desc)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "forum '" + slug + "' not found",
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
		}
		return
	}
	response := []byte("[]")
	if len(threads) != 0 {
		response, _ = threads.MarshalJSON()
	}
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
