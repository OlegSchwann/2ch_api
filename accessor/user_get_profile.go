package accessor

import (
	"2ch_api/types"
	"github.com/jackc/pgx"
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

// TODO: зарефакторить - всю обработку специфичных ошибок внести сюда,
// возвращать accessor.Error
func (cp *ConnPool) UserGetProfile(nickname string) (user types.User, err error) {
	err = cp.ConnPool.QueryRow("UserGetProfile", nickname).Scan(
		&user.Nickname, &user.About, &user.Email, &user.Fullname)
	return
}