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
  "nickname",
  "about",
  "email",
  "fullname"
from
  "user"
where 
  "nickname" = $1
;`
		_, err = conn.Prepare("UserGetProfile", sql)
		return
	})
}

func (cp *ConnPool) UserGetProfile(nickname string) (user types.User, err error) {
	err = cp.ConnPool.QueryRow("UserGetProfile", nickname).Scan(
		&user.Nickname, &user.About, &user.Email, &user.Fullname)
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
