package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) UserUpdateProfile(ctx *fasthttp.RequestCtx) {
	requestUser := types.UserUpdate{}
	err := requestUser.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusBadRequest)
		return
	}
	requestUser.Nickname = ctx.UserValue("nickname").(string)
	err, responseUser, status := e.ConnPool.UserUpdateProfile(requestUser)
	switch status {
	case accessor.StatusOk:
		response, _ := responseUser.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusOK)
		return
	case accessor.StatusConflict:
		response, _ := types.Error{
			Message: "Conflict with other user: '" + requestUser.Nickname + "'",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusConflict)
		return
	case accessor.StatusNotFound:
		response, _ := types.Error{
			Message: "Can not found user '" + requestUser.Nickname + "'",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
		return
	case accessor.StatusInternalServerError:
		response, _ := types.Error{
			Message: "Internal serer error: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
}
