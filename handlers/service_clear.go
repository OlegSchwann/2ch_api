package handlers

import (
	"2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
)

// truncate table для всех таблиц, быстро уничтожает данные.
func (e *Environment)ServiceClear(ctx *fasthttp.RequestCtx) {
	err := e.ConnPool.ServiceClear()
	if err != nil {
		response, _ := types.Error{
			Message: "unable to truncate table: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
