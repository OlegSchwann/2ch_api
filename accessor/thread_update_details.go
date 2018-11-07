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
  "thread"
set
  "message" = $2,
  "title" = $3
where 
  "id" = $1
;`
		_, err = conn.Prepare("ThreadUpdateDetailsUpdateMessageTitle", sql)
		return err
	})
}

func (cp *ConnPool) ThreadUpdateDetailsUpdateMessageTitle(id int, message string, title string) (err error) {
	commandTag, err := cp.Exec("ThreadUpdateDetailsUpdateMessageTitle", id, message, title)
	if commandTag.RowsAffected() == 0 {
		err = &Error{
			Code: http.StatusNotFound,
			UnderlyingError: err,
		}
	}
	return
}
