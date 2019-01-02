package accessor

import (
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "author",
  "created",
  "forum",
  "id",
  "message",
  "slug",
  "title",
  "number_of_children"
from
  "thread"
where
  "id" = $1
;`
		_, err = conn.Prepare("PostCreateGetThreadById", sql)
		return err
	})
}

func (cp *ConnPool) PostCreateGetThreadById(id int) (thread types.Thread, err error) {
	err = cp.QueryRow("PostCreateGetThreadById", id).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.NumberOfChildren)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "author",
  "created",
  "forum",
  "id",
  "message",
  "slug",
  "title",
  "number_of_children"
from
  "thread"
where
  "slug" = $1
;`
		_, err = conn.Prepare("PostCreateGetThreadBySlug", sql)
		return err
	})
}

func (cp *ConnPool) PostCreateGetThreadBySlug(slug string) (thread types.Thread, err error) {
	err = cp.QueryRow("PostCreateGetThreadBySlug", slug).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.NumberOfChildren)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
	}
	return
}

// асоциативный массив родительских постов.
// MaterializedPath нового = MaterializedPath + "." + zeroPaddedNum(NumberOfChildren)
// надо обновить NumberOfChildren++
type PostConnections map[int] postConnectionsItem // ключ = types.Post.Id
type postConnectionsItem struct{
	ThreadId         int    // types.Post.ThreadId
	MaterializedPath string // types.Post.MaterializedPath
	NumberOfChildren int    // types.Post.NumberOfChildren
}

func (cp *ConnPool) PostCreateGetParentPosts(parents []int) (postConnections PostConnections, err error) {
	buffer := strings.Builder{}
	buffer.Write([]byte(`
select
  "id",
  "thread_id",
  "materialized_path",
  "number_of_children"
from
  "post"
where id in (`))
	parentsAsString := make([]string, len(parents))
	for i, v := range parents {
		parentsAsString[i] = strconv.Itoa(v)
	}
	buffer.Write([]byte(strings.Join(parentsAsString, ",")))
	buffer.Write([]byte(");"))

	rows, err := cp.Query(buffer.String())
	defer rows.Close()
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}

	postConnections = make(PostConnections)
	for rows.Next() {
		var id int
		var pc postConnectionsItem
		err = rows.Scan(&id, &pc.ThreadId, &pc.MaterializedPath, &pc.NumberOfChildren)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		postConnections[id] = pc
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update
  "thread"
set
  "number_of_children" = $1
where
  "id" = $2
;`
		_, err = conn.Prepare("PostCreateUpdateThreadNumberOfChildren", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update
  "post"
set
  "number_of_children" = $1
where "id" = $2
;`
		_, err = conn.Prepare("PostCreateUpdatePostNumberOfChildren", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "post" (
  "thread_id",
  "author",            
  "created",
  "message",           
  "parent",
  "forum",
  "thread_slug",
  "materialized_path"
) values (
  $1,  	
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
) returning
  "author",
  "created",
  "forum",
  "id",
  "message",
  "parent",
  "thread_id"
;`
		_, err = conn.Prepare("PostCreateInsertOnePost", sql)
		return err
	})
}

func (cp *ConnPool) PostsCreateInsert(
	tread types.Thread, parentPost PostConnections, posts types.Posts) (
	responsePosts types.Posts, err error) {
	// всё добавление постов в одну транзакцию.

	tx, err := cp.Begin()
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}

	defer func() { // необходимо быть уверенным, что транзакция завершится.
		if err != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.Wrap(err, txErr.Error())
			}
		} else {
			err = tx.Commit()
		}
		return
	}()

	_, err = tx.Exec("PostCreateUpdateThreadNumberOfChildren", tread.NumberOfChildren, tread.Id)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	for key, value := range parentPost {
		_, err = tx.Exec("PostCreateUpdatePostNumberOfChildren", value.NumberOfChildren, key)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
	}
	// TODO: переписать на batch.
	for _, post := range posts {
		responsePost := types.Post{}
		err = tx.QueryRow(
			"PostCreateInsertOnePost",
			post.ThreadId,
			post.Author,
			post.Created,
			post.Message,
			post.Parent,
			post.Forum,
			post.ThreadSlug,
			post.MaterializedPath).Scan(
			&responsePost.Author,
			&responsePost.Created,
			&responsePost.Forum,
			&responsePost.Id,
			&responsePost.Message,
			&responsePost.Parent,
			&responsePost.ThreadId)
		if err != nil {
			pgxPgError := err.(pgx.PgError)
			if pgxPgError.Code == "23503" {
				// Insert or update on table "post" violates foreign key constraint "post_author_fkey"
				err = &Error{
					Code:            http.StatusNotFound,
					UnderlyingError: err,
				}
			} else {
				err = &Error{
					Code:            http.StatusInternalServerError,
					UnderlyingError: err,
				}
			}
			return
		}
		responsePosts = append(responsePosts, responsePost)
	}
	return
}

/* Структура для быстрого нахождения дерева коммкетариев:
┌─────────┬──────────────────────┬────────┐
│структура│Материализованный путь│потомков│
├─────────┼──────────────────────┼────────┤
│ a       │ 01                   │ 2      │
│ -b      │ 01.01                │ 1      │
│ --c     │ 01.01.01             │ 0      │
│ -d      │ 01.02                │ 3      │
│ --e     │ 01.02.01             │ 1      │
│ ---f    │ 01.02.01.01          │ 0      │
│ --g     │ 01.02.02             │ 1      │
│ ---h    │ 01.02.02.01          │ 0      │
│ --$     │ 01.02.03             │ 0      │
│ i       │ 02                   │ 1      │
│ -j      │ 02.01                │ 0      │
│ k       │ 03                   │ 1      │
│ -l      │ 03.01                │ 0      │
│ m       │ 04                   │ 1      │
│ -n      │ 04.01                │ 3      │
│ --o     │ 04.01.01             │ 1      │
│ ---p    │ 04.01.01.01          │ 1      │
│ ----q   │ 04.01.01.01.01       │ 1      │
│ -----r  │ 04.01.01.01.01.01    │ 1      │
│ ------s │ 04.01.01.01.01.01.01 │ 1      │
│ -t      │ 04.02                │ 2      │
│ --u     │ 04.02.01             │ 1      │
│ ---v    │ 04.02.01.01          │ 0      │
│ --w     │ 04.02.02             │ 0      │
│ -x      │ 04.03                │ 0      │
│ y       │ 05                   │ 0      │
│ z       │ 06                   │ 0      │
└─────────┴──────────────────────┴────────┘
*/

// new (вставляемая строка) гарантированно содержит:
//   "thread_id" или "thread_slug"
//   "author"
//   "created" -- вставляется на стороне application server
//   "message"
//   "parent" -- может быть равно 0, если сообщение корневое
//
// не нуждаются в триггере
//   "id" -- добавляется самостоятельно из serial
//   "is_edited" -- по умолчанию false
//
// поддерживаются триггером:
//   "forum" -- берётся из "thread"."forum"
//   "materialized_path"
//   "number_of_children" -- важно обновить для родительского сообщения.

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "user_in_forum"(
  "forum",
  "nickname"
) values (
  $1,
  $2
);`
		_, err = conn.Prepare("InsertIntoUserInForum", sql)
		return err
	})
}

// При добавлении в posts добавляем пользоватля в список пользователей этого форума.
// Просто пытаемся добавить автора поста к списку пользователей этого форума.
// Ошибку не уникальности не обрабатываем.
func (cp *ConnPool) InsertIntoUserInForum(forum string, nickname string)(err error){
	_, err = cp.Exec("InsertIntoUserInForum", forum, nickname)
	return
}
