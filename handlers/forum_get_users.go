package handlers

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ForumGetUsers(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	queryArgs := ctx.Request.URI().QueryArgs()
	limit, _ := queryArgs.GetUint("limit")
	since := string(queryArgs.Peek("since"))
	desc := queryArgs.GetBool("desc")

	// TODO: logic
	fmt.Print(slug, limit, since, desc)

	if true {
		responseUsers := types.Users{}
		response, _ := responseUsers.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusOK)
	} else {
		response, _ := types.Error{
			Message: "Can't find user with id '" + slug + "'.",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
	}
}
