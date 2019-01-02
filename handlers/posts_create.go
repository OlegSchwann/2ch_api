package handlers

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"time"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/shared_helpers"
	"github.com/OlegSchwann/2ch_api/types"
)

func (e *Environment) PostsCreate(ctx *fasthttp.RequestCtx) {
	// вытаскиваем из запроса все возможные данные.
	posts := types.Posts{}
	if err := posts.UnmarshalJSON(ctx.Request.Body()); err != nil {
		response, _ := types.Error{
			Message: "unable to parse json: " + err.Error(),
		}.MarshalJSON()
		ctx.Write(response)
		ctx.Response.Header.SetStatusCode(http.StatusUnprocessableEntity)
		return
	}

	// вытаскиваем связанный thread, отдаём 404, если такого нет.
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

	// Отдаём сразу 201, если требуется записать пустой массив постов.
	if len(posts) == 0 {
		ctx.WriteString("[]")
		ctx.Response.Header.SetStatusCode(http.StatusCreated)
		return
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
	// Горутина считывает количество детей, формирует материализованные пути и при записи новых
	// строк увеличивает количество детей. Если кто-то другой считает количество в промежутке,
	// и попытается записать материализованные пути, то упадём с ошибкой уникальности
	// материализованного пути. Это произойдёт, если только если попытаться добавить ответы к
	// одному узлу одновременно. Пессимистичное решение - блокировать в момент чтения строку с
	// постом или тредом. Оптимистичное решение - в случае падения пересчитывать
	// материализованные пути. (spin lock, пробовать вставить, поа не получится.) Пока проходит
	// тесты при заполнении, не падает. Сейчас считаем, что к одному узлу в разных потоках едва ли
	// будут добавлять, применяем оптимистичное решение.

	// собираем "materialized_path" для вставляемых постов, обновляем "number_of_children" у родителей.
	{
		for i := range posts {
			if posts[i].Parent == 0 { // если родитель - thread
				thread.NumberOfChildren ++
				posts[i].MaterializedPath = shared_helpers.ZeroPad(uint(thread.NumberOfChildren), 6)
			} else { // если родитель - post
			    parentId := posts[i].Parent
				parentPostConnections := postsConnections[parentId]

				// проверяем условие, что родитель находится в той же ветке обсуждения, что и сам пост.
				if parentPostConnections.ThreadId != posts[i].ThreadId {
					response, _ := types.Error{
						Message: "Parent post was created in another thread: ",
					}.MarshalJSON()
					ctx.Write(response)
					ctx.Response.Header.SetStatusCode(http.StatusConflict)
					return
				}
				parentPostConnections.NumberOfChildren ++
				posts[i].MaterializedPath = parentPostConnections.MaterializedPath + "." +
					shared_helpers.ZeroPad(uint(parentPostConnections.NumberOfChildren), 6)
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
			ctx.Response.Header.SetStatusCode(http.StatusNotFound)
			return
		}
		if accessorError.Code == http.StatusInternalServerError {
			response, _ := types.Error{
				Message: "Server error: " + err.Error(),
			}.MarshalJSON()
			ctx.Write(response)
			ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
			return
		}
	}

	// Делаем то, что раньше делалось триггером -
	// в таблицу "user_in_forum" добавляем всех пользователей - авторов постов.
	for _, post := range posts {
		e.ConnPool.InsertIntoUserInForum(post.Forum, post.Author)
	}
	
	response, _ := responsePosts.MarshalJSON()
	ctx.Write(response)
	ctx.Response.Header.SetStatusCode(http.StatusCreated)
	return
}
