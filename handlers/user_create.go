package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) UserCreate(ctx *fasthttp.RequestCtx) {
	requestUser := types.User{}
	err := requestUser.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "Unable unmarshal json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusBadRequest)
		return
	}
	requestUser.Nickname = ctx.UserValue("nickname").(string)

	err, conflictedUsers := e.ConnPool.UserCreate(requestUser)
	if err != nil {
		response, _ := types.Error{
			Message: "server error: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	if len(conflictedUsers) != 0 {
		response, _ := conflictedUsers.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusConflict)
		return
	}
	response, _ := requestUser.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
}
