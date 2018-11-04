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
  "id" = $1
;`
		_, err = conn.Prepare("PostGetDetailsSelectPost", sql)
		return err
	})
}

func (cp *ConnPool) PostGetDetailsSelectPost(postId int) (post types.Post, err error) {
	err = cp.QueryRow("PostGetDetailsSelectPost", postId).Scan(
		&post.ThreadId, &post.Author, &post.Created, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Forum)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code: http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}
