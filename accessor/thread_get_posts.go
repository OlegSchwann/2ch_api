package accessor

import (
	"github.com/OlegSchwann/2ch_api/shared_helpers"
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"
)

// flat сортировка

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"  
from
  "post"
where 
  "thread_id" = $1
order by
  ("thread_id", "id") asc
limit $2
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSortAsc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1
order by
  ("thread_id", "id") desc
limit $2
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSortDesc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsFlatSort(threadId int, limit int, desc bool) (posts types.Posts, err error) {
	rows := (*pgx.Rows)(nil)
	if desc {
		rows, err = cp.Query("ThreadGetPostsFlatSortDesc", threadId, limit)
	} else {
		rows, err = cp.Query("ThreadGetPostsFlatSortAsc", threadId, limit)
	}
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1 and
  "id" > $2
order by
  ("thread_id", "id") asc -- как в индексе указано.
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSinceSortAsc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1 and
  "id" < $2
order by
  ("thread_id", "id") desc
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSinceSortDesc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsFlatSince(threadId int, limit int, since int, desc bool) (posts types.Posts, err error) {
	rows := (*pgx.Rows)(nil)
	if desc {
		rows, err = cp.Query("ThreadGetPostsFlatSinceSortDesc", threadId, since, limit)
	} else {
		rows, err = cp.Query("ThreadGetPostsFlatSinceSortAsc", threadId, since, limit)
	}
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

// tree сортировка. Комментарии выводятся согласно их уровням и пагинируются последовательно.
// limit распространяется просто на количество выведенных постов, абсолютно не важно какая вложенность. Пример limit 3:
//       asc               desc
// │page 1│page 2│   │page 1│page 2│
// │ 1    │ 1.3  │   │ 2    │ 1.2  │
// │ 1.1  │ 2    │   │ 1    │ 1.3  │
// │ 1.2  │      │   │ 1.1  │      │

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1
order by
  ("thread_id", "materialized_path") desc
limit $2
;`
		_, err = conn.Prepare("ThreadGetPostsTreeSortDesc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1
order by
  ("thread_id", "materialized_path") asc
limit $2
;`
		_, err = conn.Prepare("ThreadGetPostsTreeSortAsc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsTree(threadId int, limit int, desc bool) (posts types.Posts, err error) {
	rows := (*pgx.Rows)(nil)
	if desc {
		rows, err = cp.Query("ThreadGetPostsTreeSortDesc", threadId, limit)
	} else {
		rows, err = cp.Query("ThreadGetPostsTreeSortAsc", threadId, limit)
	}
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1 and 
  "materialized_path" < (select
    "materialized_path"
  from
    "post"
  where
    "id" = $2
  )
order by
  ("thread_id", "materialized_path") desc
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsTreeSortSinceDesc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post"
where 
  "thread_id" = $1 and
  "materialized_path" > (select
    "materialized_path"
  from
    "post"
  where
    "id" = $2
  )
order by
  ("thread_id", "materialized_path") asc
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsTreeSinceSortAsc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsTreeSince(threadId int, limit int, since int, desc bool) (posts types.Posts, err error) {
	rows := (*pgx.Rows)(nil)
	if desc {
		rows, err = cp.Query("ThreadGetPostsTreeSortSinceDesc", threadId, since, limit)
	} else {
		rows, err = cp.Query("ThreadGetPostsTreeSinceSortAsc", threadId, since, limit)
	}
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

// parent_tree сортировка. Комментарии выводятся согласно их уровням, но пагинируются по корневым. Пример limit 3
// Выводим всех детей. При сортировке DESC мы в таком порядке сортируем только первое число в пути поста,
// остальной путь ВСЕГДА по умолчанию будет отсортирован при помощи ASC.
//       asc              desc
// │page 1│page 2│   │page 1│page 2│
// │ 1    │ 4    │   │ 4    │ 1    │
// │ 1.1  │ 4.1  │   │ 4.1  │ 1.1  │
// │ 1.2  │      │   │ 3    │ 1.2  │
// │ 2    │      │   │ 3.1  │      │
// │ 2.1  │      │   │ 2    │      │
// │ 2.1.1│      │   │ 2.1  │      │
// │ 3    │      │   │ 2.1.1│      │
// │ 3.1  │      │   │      │      │

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select 
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum",
  "materialized_path"  
from
  "post"
where 
  "thread_id" = $1 and 
  "materialized_path" > $2 and 
  "materialized_path" < $3
order by 
  ("thread_id", "materialized_path") asc
;`
		_, err = conn.Prepare("ThreadGetPostsParentTreeSortAsc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsParentTreeSortAsc(threadId int, limit int) (
	posts types.Posts, err error) {

	materializedPathUp := ""
	materializedPathDown := ""
	// материализованный путь < пути первого поста, когда не задан since
	materializedPathUp = "000000"
	// материализованный путь первого поста, когда не задан since
	materializedPathDown = shared_helpers.ZeroPad(uint(limit+1), 6)
	rows, err := cp.Query("ThreadGetPostsParentTreeSortAsc", threadId, materializedPathUp, materializedPathDown)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum, &post.MaterializedPath)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select 
  "materialized_path"
from
  "post"
where 
  "id" = $1
;`
		_, err = conn.Prepare("ThreadGetPostsGetMaterializedPathById", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsParentTreeSinceSortAsc(threadId int, limit int, since int) (
	posts types.Posts, err error) {

	materializedPathUp := ""
	err = cp.QueryRow("ThreadGetPostsGetMaterializedPathById", since).Scan(&materializedPathUp)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	pathUp, _ := strconv.Atoi(materializedPathUp[0:6]) // вырезаем первое число: "0͟0͟0͟0͟0͟3.000001"
	materializedPathDown := shared_helpers.ZeroPad(uint((limit+1)+pathUp), 6)
	rows, err := cp.Query("ThreadGetPostsParentTreeSortAsc", threadId, materializedPathUp, materializedPathDown)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum, &post.MaterializedPath)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

// страшная функция на substrin.
// TODO: проверить работоспособность и в случае чего отделить префикс материализованного пути в отдельную таблицу.
func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select 
  "number_of_children"
from
  "thread"
where 
  "id" = $1
;`
		_, err = conn.Prepare("ThreadGetPostsGetThreadLastMaterializedPath", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
with "bounds" as (select 
  substring($2 from 1 for 6) as "from_prefix",
  substring($2 from 8      ) as "from_postfix",
  substring($3 from 1 for 6) as "to_prefix",
  substring($3 from 8      ) as "to_postfix"
)(
select  
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post",
  "bounds"
where
  "thread_id" = $1 and 
  substring("materialized_path" from 1 for 6) = bounds.from_prefix and 
  substring("materialized_path" from 8      ) < bounds.from_postfix
order by
  "thread_id", substring("materialized_path" from 1 for 6) desc, substring("materialized_path" from 8) asc
) union all (
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post",
  "bounds"
where 
  "thread_id" = $1 and 
  substring("materialized_path" from 1 for 6) > bounds.from_prefix and 
  substring("materialized_path" from 1 for 6) < bounds.to_prefix
order by
  "thread_id", substring("materialized_path" from 1 for 6) desc, substring("materialized_path" from 8) asc
) union all (
select
  "thread_id",
  "author",
  "created",
  "id", 
  "is_edited",
  "message",
  "parent",
  "forum"
from
  "post",
  "bounds"
where
  "thread_id" = $1 and
  substring("materialized_path" from 1 for 6) = bounds.to_prefix and
  substring("materialized_path" from 8) > bounds.to_postfix
order by
  "thread_id", substring("materialized_path" from 1 for 6) desc, substring("materialized_path" from 8) asc
);`
		_, err = conn.Prepare("ThreadGetPostsParentTreeSortDesc", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsParentTreeSortDesc(threadId int, limit int) (
	posts types.Posts, err error) {


	pathDown := 0
	err = cp.QueryRow("ThreadGetPostsGetThreadLastMaterializedPath", threadId).Scan(&pathDown)
	if err != nil {
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	materializedPathDown := shared_helpers.ZeroPad(uint(pathDown+1), 6)
	pathUp := pathDown - limit
	if pathUp < 0 {
		pathUp = 0
	}
	materializedPathUp := shared_helpers.ZeroPad(uint(pathUp), 6)

	rows, err := cp.Query("ThreadGetPostsParentTreeSortDesc", threadId, materializedPathUp, materializedPathDown)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

func (cp *ConnPool) ThreadGetPostsParentTreeSinceSortDesc(threadId int, limit int, since int) (
	posts types.Posts, err error) {

	materializedPathDown := ""
	err = cp.QueryRow("ThreadGetPostsGetMaterializedPathById", since).Scan(&materializedPathDown)
	if err != nil {
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}

	pathDown, _ := strconv.Atoi(materializedPathDown[0:6])
	pathUp := pathDown - limit
	if pathUp < 0 {
		pathUp = 0
	}
	materializedPathUp := shared_helpers.ZeroPad(uint(pathUp), 6)

	rows, err := cp.Query("ThreadGetPostsParentTreeSortDesc", threadId, materializedPathUp, materializedPathDown)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
		if err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		posts = append(posts, post)
	}
	return
}

// проверка существования thread.
func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select exists(
 select
   true 
 from
   "thread"
 where
   "id" = $1
);`
		_, err = conn.Prepare("ThreadGetPostsCheckIfThreadExists", sql)
		return err
	})
}

func (cp *ConnPool) ThreadGetPostsCheckIfThreadExists(threadId int) (threadExists bool, err error) {
	err = cp.QueryRow("ThreadGetPostsCheckIfThreadExists", threadId).Scan(&threadExists)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}
