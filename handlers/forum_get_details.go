package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ForumGetDetails(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	forum, err := e.ConnPool.ForumGetDetails(slug)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Can't found forum '" + slug + "': " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		if accessorError.Code == http.StatusInternalServerError {
			response, _ := types.Error{
				Message: "Unexpected error: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	response, err := forum.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
