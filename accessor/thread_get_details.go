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
		_, err = conn.Prepare("ThreadGetDetailsBySlug", sql)
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
  "id" = $1
;`
		_, err = conn.Prepare("ThreadGetDetailsById", sql)
		return
	})
}

func (cp *ConnPool) ThreadGetDetailsBySlug(slag string) (thread types.Thread, err error) {
	err = cp.QueryRow("ThreadGetDetailsBySlug", slag).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
		}
	}
	return
}

func (cp *ConnPool) ThreadGetDetailsById(id int) (thread types.Thread, err error) {
	err = cp.QueryRow("ThreadGetDetailsById", id).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
		}
	}
	return
}
