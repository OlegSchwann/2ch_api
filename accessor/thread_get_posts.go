package accessor

import (
	"github.com/jackc/pgx"
	"net/http"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select 
  "id"
from
  "thread"
where 
  slug = $1
;`
		_, err = conn.Prepare("ThreadGetPostsIdBySlug", sql)
		return
	}) // я знаю, что код дублируется. Пока поддерживаю независимость handler'ов друг от друга для простоты.
}

// Превращаем "thread"."slug" в "thread"."id" .
func (cp *ConnPool) ThreadGetPostsIdBySlug(threadSlug string) (threadId int, err error) {
	err = cp.QueryRow("ThreadGetPostsIdBySlug", threadSlug).Scan(&threadId)
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
  "thread_id"
from
  "post"
where 
  "thread_id" = $1 and
  "thread_id" > $2
order by
  (thread_id, created) asc -- как в индексе указано.
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSortAsc", sql)
		return err
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
  "thread_id"
from
  "post"
where 
  "thread_id" = $1 and
  "thread_id" > $2
order by
  (thread_id, created) asc -- как в индексе указано.
limit $3
;`
		_, err = conn.Prepare("ThreadGetPostsFlatSortAsc", sql)
		return err
	})

}
