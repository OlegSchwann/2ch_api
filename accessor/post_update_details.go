package accessor

import (
	"github.com/jackc/pgx"
	"net/http"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update
  "post"
set
  "message" = $2,
  "is_edited" = true
where 
  "id" = $1 -- and "message" is distinct from $2 -- нельзя использовать, нельзя отличить не найдено от не изменилось. 
;`
		_, err = conn.Prepare("PostUpdateDetailInsertMessage", sql)
		return err
	})
}

func (cp *ConnPool) PostUpdateDetailInsertMessage(postId int, message string) (err error) {
	commandTag, err := cp.Exec("PostUpdateDetailInsertMessage", postId, message)
	if commandTag.RowsAffected() == 0 {
		err = &Error{
			Code:            http.StatusNotFound,
			UnderlyingError: err,
		}
	}
	return
}
