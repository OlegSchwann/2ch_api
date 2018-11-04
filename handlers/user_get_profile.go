package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) UserGetProfile(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)
	user, err := e.ConnPool.UserGetProfile(nickname)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Can't find user with nickname '" + nickname + "'.",
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		response, _ := types.Error{
			Message: "Internal server error: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	response, _ := user.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}
