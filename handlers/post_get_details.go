package handlers

import (
	"2ch_api/types"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"strings"
)

func (e *Environment)PostGetDetails(ctx *fasthttp.RequestCtx) {
	idString := ctx.UserValue("id").(string)
	id, _ := strconv.Atoi(idString)

	related := string(ctx.URI().QueryArgs().Peek("related"))
	fullInformation := struct {
		User   bool
		Forum  bool
		Thread bool
	}{
		strings.Contains(related, "user"),
		strings.Contains(related, "forum"),
		strings.Contains(related, "thread"),
	}

	// TODO: logic
	fmt.Print(id, fullInformation)

	if true {
		responsePostFull := types.PostFull{}
		response, _ := responsePostFull.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusOK)
	} else {
		response, _ := types.Error{
			Message: "Can't find Post with id '" + idString + "'.",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusNotFound)
	}
}

