package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

func (e *Environment)ThreadUpdateDetails(ctx *fasthttp.RequestCtx) {
	slugOrId := ctx.UserValue("slug_or_id").(string)

	threadUpdate := types.ThreadUpdate{}
	err := threadUpdate.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "unrecognized request: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}

	var responseThread types.Thread
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
				Message: "ThreadId '" + slugOrId + "' not found: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
		}
		return
	}

	if threadUpdate.Message == nil && threadUpdate.Title == nil {
		response, _ := responseThread.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusOK)
		return
	}

	if threadUpdate.Message == nil {
		threadUpdate.Message = &responseThread.Message
	} else {
		responseThread.Message = *threadUpdate.Message
	}

	if threadUpdate.Title == nil {
		threadUpdate.Title = &responseThread.Title
	} else {
		responseThread.Title = *threadUpdate.Title
	}

	err = e.ConnPool.ThreadUpdateDetailsUpdateMessageTitle(responseThread.Id, *threadUpdate.Message, *threadUpdate.Title)

	response, _ := responseThread.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}
