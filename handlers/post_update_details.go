package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

// Извлекаем message и id.
// запрашиваем предыдущее значение.
// если not found - возвращаем 404
// если message равно - возвращаем такое же и 200
// иначе обновляем и возвращаем новое.

func (e *Environment) PostUpdateDetails(ctx *fasthttp.RequestCtx) {
	postStringId := ctx.UserValue("id").(string)
	postId, err := strconv.Atoi(postStringId)
	//region if err != nil UnprocessableEntity
	if err != nil {
		response, _ := types.Error{
			Message: "post identificator '" + postStringId + "' must be integer",
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
		return
	} //endregion
	postUpdate := types.PostUpdate{}
	err = postUpdate.UnmarshalJSON(ctx.Request.Body())
	//region if err != nil UnprocessableEntity
	if err != nil {
		response, _ := types.Error{
			Message: "unable to parse json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
		return
	} //endregion

	post, err := e.ConnPool.PostGetDetailsSelectPost(postId)
	//region if err != nil NotFound
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "can not find post '" + postStringId + "': " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	} //endregion

	// ничего не изменяем при пустом объекте на входе.
	if postUpdate.Message == nil {
		response, _ := post.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusOK)
		return
	}

	// Ничего не изменяем, если сообщение совпадает с исходным.
	if *postUpdate.Message == post.Message {
		response, _ := post.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusOK)
		return
	}
	err = e.ConnPool.PostUpdateDetailInsertMessage(postId, *postUpdate.Message)
	//region if err != nil NotFound
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	} //endregion

	post.Message = *postUpdate.Message
	post.IsEdited = true

	response, _ := post.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
