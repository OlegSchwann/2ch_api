package handlers

import (
	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func zeroPad(integer uint, overallLen int) string {
	num := strconv.FormatUint(uint64(integer), 10)
	return strings.Repeat("0", overallLen-len(num)) + num
}

func (e *Environment) PostsCreate(ctx *fasthttp.RequestCtx) {
	// вытаскиваем из запроса все возможные данные.
	posts := types.Posts{}
	if err := posts.UnmarshalJSON(ctx.Request.Body()); err != nil {
		response, _ := types.Error{
			Message: "unable to parse json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
		return
	}
	// вытаскиваем связанный thread, отдаём 404 если нет подобного, и сразу 201 если передан пустой массив.
	thread := types.Thread{}
	{
		slugOrId := ctx.UserValue("slug_or_id").(string)
		id, err := strconv.Atoi(slugOrId)
		if err == nil {
			thread, err = e.ConnPool.PostCreateGetThreadById(id)
		} else {
			thread, err = e.ConnPool.PostCreateGetThreadBySlug(slugOrId)
		}
		if err != nil {
			accessorError := err.(*accessor.Error)
			if accessorError.Code == http.StatusNotFound {
				response, _ := types.Error{
					Message: "unable find forum '" + slugOrId + "' : " + err.Error(),
				}.MarshalJSON()
				ctx.Write(response)
				ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
				ctx.Response.Header.SetStatusCode(http.StatusNotFound)
				return
			}
			response, _ := types.Error{
				Message: err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	{
		if len(posts) == 0 {
			ctx.Write([]byte("[]"))
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusCreated)
			return
		}
	}

	// собираем информацию о родительских постах.
	// Post.Id
	// Post.ThreadId
	// Post.MaterializedPath
	// Post.NumberOfChildren
	var postsConnections accessor.PostConnections
	{
		// проставляем поля для post, которые нам известны из thread.
		timeNow := time.Now()
		// собираем все ссылки на родителей в массив
		parentIds := make([]int, len(posts))
		for i := range posts {
			posts[i].Created = timeNow
			posts[i].ThreadId = thread.Id
			posts[i].ThreadSlug = thread.Slug
			posts[i].Forum = thread.Forum

			if posts[i].Parent != 0 {
				parentIds = append(parentIds, posts[i].Parent)
			}
		}
		// вытаскиваем описания родительских постов
		err := error(nil)
		postsConnections, err = e.ConnPool.PostCreateGetParentPosts(parentIds)
		if err != nil {

		}
	}
	// Тут гонка данных.
	// Горутина считывает количество детей, формирует материализованные пути и при записи новых строк увеличивает количество детей.
	// если кто-то другой считает количество в промежутке, и попытается записать материализованные пути, то упадём с ошибкой уникальности материализованного путию
	// это произойдёт, если только если попытаться добавить ответы к одному узлу одновременно.
	// пессимистичное решение - блокировать в момент чтения строку с постом или тредом.
	// оптимистичное решение - в случае падения пересчитывать материализованные пути. (spin lock, пробовать вставить, поа не получится.)
	// сейчас считаем, что к одному узлу в разных потоках едва ли будут добавлять, применяем оптимистичное решение.

	// собираем "materialized_path" для вставляемых постов, обновляем "number_of_children" у родителей.
	{
		for i := range posts {
			if posts[i].Parent == 0 { // если родитель - thread
				thread.NumberOfChildren ++
				posts[i].MaterializedPath = zeroPad(uint(thread.NumberOfChildren), 6)
			} else { // если родитель - post
			    parentId := posts[i].Parent
				parentPostConnections := postsConnections[parentId]

				// проверяем условие, что родитель находится в той же ветке обсуждения, что и сам пост.
				if parentPostConnections.ThreadId != posts[i].ThreadId {
					response, _ := types.Error{
						Message: "Parent post was created in another thread: ",
					}.MarshalJSON()
					ctx.Write(response)
					ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
					ctx.Response.Header.SetStatusCode(http.StatusConflict)
					return
				}
				parentPostConnections.NumberOfChildren ++
				posts[i].MaterializedPath = parentPostConnections.MaterializedPath + "." +
					zeroPad(uint(parentPostConnections.NumberOfChildren), 6)
				postsConnections[parentId] = parentPostConnections
			}
		}
	}

	// теперь все данные собраны, надо вставить в базу, за одну транзакцию
	// обновить количество детей у треда, родительских постов и вставить сами посты.
	responsePosts, err := e.ConnPool.PostsCreateInsert(thread, postsConnections, posts)
	if err != nil { // TODO: обработка гонки данных при сборке materialised_path, логика retry.
		accessorError := err.(*accessor.Error)
		if accessorError.Code == http.StatusNotFound {
			response, _ := types.Error{
				Message: "Not found: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		if accessorError.Code == http.StatusInternalServerError {
			response, _ := types.Error{
				Message: "Server error: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}
	response, _ := responsePosts.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
	return
}
