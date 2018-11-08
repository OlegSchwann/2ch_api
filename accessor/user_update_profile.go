package accessor

import (
	"github.com/jackc/pgx"

	"github.com/OlegSchwann/2ch_api/types"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update
  "user"
set 
  "about"    = case when $2 = '' then "about"    else $2 end,
  "email"    = case when $3 = '' then "email"    else $3 end,
  "fullname" = case when $4 = '' then "fullname" else $4 end
where 
  "nickname" = $1
returning 
  "nickname",
  "about",
  "email",
  "fullname"
;`
		_, err = conn.Prepare("UserUpdateProfile", sql)
		return err
	})
}

const (
	StatusOk = iota
	StatusConflict
	StatusNotFound
	StatusInternalServerError
)

func (cp *ConnPool) UserUpdateProfile(user types.UserUpdate) (
	err error, responseUser types.User, status int) {
	err = cp.QueryRow("UserUpdateProfile",
		user.Nickname, user.About, user.Email, user.Fullname).Scan(
		&responseUser.Nickname, &responseUser.About, &responseUser.Email, &responseUser.Fullname)
	if err != nil {
		if err.Error() == "no rows in result set" {
			status = StatusNotFound
			return
		}
		if err.(pgx.PgError).Code == "23505" { // duplicate key value violates unique constraint
			status = StatusConflict
			return
		}
		status = StatusInternalServerError
		return
	}
	status = StatusOk
	return
}
