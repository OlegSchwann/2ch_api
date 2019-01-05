package accessor

import (
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/jackc/pgx"
	"net/http"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "thread"(
  "author",
  "created",
  "forum",
--"id" will be set automatically
  "message",
  "slug",
  "title"
--"votes" are store in separate table  
) values (
  (select
    "nickname"
  from
    "user"
  where
    "nickname" = $1),
  $2,
  (select
    "slug"
  from
    "forum"
  where
    "slug" = $3),
  $4,
  $5,
  $6
) returning
  "author",
  "created",
  "forum",
  "id",
  "message",
  "slug",
  "title",
  0
;`
		_, err = conn.Prepare("ThreadCreate", sql)
		return
	})
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
  "title"
from
  "thread"
where 
  "slug" = $1
;`
		_, err = conn.Prepare("ThreadCreateConflict", sql)
		return
	})
}

func (cp *ConnPool) ThreadCreate(thread types.Thread) (realThread types.Thread, err error) {
	err = cp.QueryRow("ThreadCreate",
		thread.Author, thread.Created, thread.Forum /*id*/, thread.Message, thread.Slug, thread.Title /*votes*/).Scan(
		&realThread.Author, &realThread.Created, &realThread.Forum, &realThread.Id, &realThread.Message, &realThread.Slug, &realThread.Title, &realThread.Votes)
	if err != nil {
		pgxPgError := err.(pgx.PgError)
		if pgxPgError.Code == "23502" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		if pgxPgError.Code == "23505" { // duplicate key value violates unique constraint "thread_slug_key"
			err = cp.QueryRow("ThreadCreateConflict", thread.Slug).Scan(
				&realThread.Author, &realThread.Created, &realThread.Forum, &realThread.Id, &realThread.Message, &realThread.Slug, &realThread.Title)
			err = &Error{
				Code:            http.StatusConflict,
				UnderlyingError: err,
			}
			return
		}
	}
	return
}
