package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) ThreadGetDetails(ctx *fasthttp.RequestCtx) {
	slugOrId := ctx.UserValue("slug_or_id").(string)
	var responseThread types.Thread
	var err error
	id, err := strconv.Atoi(slugOrId)
	if err == nil {
		responseThread, err = e.ConnPool.ThreadGetDetailsById(id)
	} else {
		responseThread, err = e.ConnPool.ThreadGetDetailsBySlug(slugOrId)
	}
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Thread '" + slugOrId + "' not found: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
		}
		return
	}
	response, _ := responseThread.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
