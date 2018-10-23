package accessor

import (
	"2ch_api/types"
	"fmt"
	"github.com/jackc/pgx"
	"net/http"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into
  "forum"("slug", "title", "user")
values (
  $1,
  $2,
  (select
    "nickname"
  from
    "user"
  where
    "user"."nickname" = $3)
) returning 
  "slug", "title", "user"
;`
		_, err = conn.Prepare("ForumCreate", sql)
		return
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
with "related_threads" as (
  select
    count(*) as "count"
  from
    "thread"
  where
    "thread"."forum" = $1	 
), "related_posts" as (
  select 
    count(*) as "count"
  from
    "post"
  where
    "forum" = $1
) select
  "related_posts"."count" as "posts",
  "slug",
  "related_threads"."count" as "threads",
  "title",
  "user"
from
  "forum", 
  "related_threads",
  "related_posts"
where
  "forum"."slug" = $1
;`
		_, err = conn.Prepare("ForumCreateOnConflict", sql)
		return
	})
}

func (cp *ConnPool) ForumCreate(forum types.Forum) (returnForum types.Forum, err error) {
	err = cp.QueryRow("ForumCreate", forum.Slug, forum.Title, forum.User).Scan(
		&returnForum.Slug, &returnForum.Title, &returnForum.User)
	if err != nil {

		fmt.Printf("\n\n%#v\n\n", err)

		pgError := err.(pgx.PgError)
		if pgError.Code == "23505" { // duplicate key value violates unique constraint
			err = &Error{
				Code:            http.StatusConflict,
				UnderlyingError: err,
			}
			return
		}
		if pgError.Code == "23503" || // insert or update on table violates foreign key constraint
			pgError.Code == "23502" { // null value in column "user" violates not-null constraint
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
	}
	return
}

func (cp *ConnPool) ForumCreateOnConflict(slug string) (forum types.Forum, err error) {
	err = cp.QueryRow("ForumCreateOnConflict", slug).Scan(
		&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	if err != nil && err.Error() == "no rows in result set" {
		err = &Error{
			Code:            http.StatusNotFound,
			UnderlyingError: err,
		}
	}
	return
}
