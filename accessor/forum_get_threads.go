package accessor

import (
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/jackc/pgx"
	"net/http"
	"time"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select exists(
  select
    true 
  from
    "forum"
  where
    "slug" = $1
);`
		_, err = conn.Prepare("ForumGetThreadsCheckForumExist", sql)
		return
	})
}

func (cp *ConnPool) ForumGetThreadsCheckForumExist(forumSlug string) (exist bool, err error) {
	err = cp.QueryRow("ForumGetThreadsCheckForumExist", forumSlug).Scan(&exist)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
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
  "title"
from 
  "thread"
where 
  "forum" = $1 and 
  "created" >= $3
order by
  "created" asc
limit
  $2
;`
		_, err = conn.Prepare("ForumGetThreadsSortedAsc", sql)
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
  "forum" = $1 and 
  "created" <= $3
order by
  "created" desc
limit
  $2
;`
		_, err = conn.Prepare("ForumGetThreadsSortedDesc", sql)
		return
	})
}

func (cp *ConnPool) ForumGetThreads(slug string, limit int, since time.Time, desc bool) (
	threads types.Threads, err error) {
	exist, err := cp.ForumGetThreadsCheckForumExist(slug)
	if err != nil {
		return
	}
	if !exist {
		err = &Error{
			Code:            http.StatusNotFound,
			UnderlyingError: nil,
		}
		return
	}
	var rows *pgx.Rows
	if desc {
		rows, err = cp.Query("ForumGetThreadsSortedDesc", slug, limit, since)
	} else {
		rows, err = cp.Query("ForumGetThreadsSortedAsc", slug, limit, since)
	}
	defer rows.Close()
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	for rows.Next() {
		if err = rows.Err(); err != nil {
			err = &Error{
				Code:            http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		thread := types.Thread{}
		rows.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title)
		threads = append(threads, thread)
	}
	return
}
