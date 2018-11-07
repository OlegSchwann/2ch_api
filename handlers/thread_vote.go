package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)


// запрашиваем, существует ли такая запись в таблице, и с каким голосом.
//                                       ┌──true──<запись существует?>──false──┐
//  ┌──true──<меняется ли голос на противоположный?>──false──┐       ⎡insert vote;       ⎤
//⎡update vote;           ⎤                [запрашиваем thread;]     ⎢thread.vote += vote⎥
//⎢thread.vote += 2 * vote⎥                                          ⎣возвращая thread;  ⎦
//⎣возвращая thread;      ⎦


// TODO: зарефакторить в одну транзакцию.
func (e *Environment) ThreadVote(ctx *fasthttp.RequestCtx) {
	vote := types.Vote{}
	err := vote.UnmarshalJSON(ctx.Request.Body())
	if err != nil {
		response, _ := types.Error{
			Message: "Can not parse json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
	}

	slugOrId := ctx.UserValue("slug_or_id").(string)
	// восстанавлиаем "thread"."id"
	threadId, err := strconv.Atoi(slugOrId)
	if err != nil {
		threadId, err = e.ConnPool.ThreadVoteGetThreadIdBySlug(slugOrId)
		if err != nil {
			accessorError := err.(*accessor.Error)
			if accessorError.Code == http.StatusNotFound {
				response, _ := types.Error{
					Message: err.Error(),
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
		}
	}

	// находим предыдущий голос этого же пользователя, если есть.
	oldVote, err := e.ConnPool.ThreadVoteSelectOldVote(vote.Nickname, threadId)
	if err != nil {
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			// если ещё небыло голоса за этот пост, добавляем голос и обновляем количество голосов.
			err = e.ConnPool.ThreadVoteInsert(vote.Nickname, vote.Voice, threadId)
			if err != nil {
				accessorError = err.(*accessor.Error)
				if accessorError.Code == http.StatusNotFound {
					response, _ := types.Error{
						Message: err.Error(),
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
			}
			thread, err := e.ConnPool.ThreadVoteUpdateThreadVote(vote.Voice, threadId)
			if err != nil {
				response, _ := types.Error{
					Message: err.Error(),
				}.MarshalJSON()
				ctx.Write(response)
				ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
				return
			}
			// возвращаем корректный код.
			response, _ := thread.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusOK)
			return
		}
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	// если пользователь повторно послал тот же голос
	if vote.Voice == oldVote {
        // то вычитываем thread, не потерявший актуальности, отдаём его.
        thread, err := e.ConnPool.ThreadVoteSelectUnchangedThread(threadId)
		if err != nil {
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
		response, _ := thread.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusOK)
		return
	}
	// если пользователь прислал противоположный запрос, то обновляем запись голоса
	err = e.ConnPool.ThreadVoteUpdateVote(vote.Nickname, vote.Voice, threadId)
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	// и изменяем сумму в "thread"."vote", сразу вытаскивая ответ
	thread, err := e.ConnPool.ThreadVoteUpdateThreadVote(vote.Voice * 2, threadId)
	if err != nil {
		response, _ := types.Error{
			Message: err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
		return
	}
	response, _ := thread.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusOK)
	return
}
