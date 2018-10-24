package accessor

import (
	"github.com/jackc/pgx"
	"net/http"

	"github.com/OlegSchwann/2ch_api/types"
)

func init() {
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
		_, err = conn.Prepare("ForumGetDetails", sql)
		return
	})
}

func (cp *ConnPool) ForumGetDetails(slag string) (forum types.Forum, err error) {
	err = cp.QueryRow("ForumGetDetails", slag).Scan(
		&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
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
		return
	}
	return
}
