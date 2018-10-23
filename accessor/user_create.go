package accessor

import (
	"2ch_api/types"
	"github.com/jackc/pgx"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "user" (
  "nickname",
  "about",
  "email",
  "fullname"
) values (
  $1,
  $2,
  $3,
  $4
);`
		_, err = conn.Prepare("UserCreate", sql)
		return
	})

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
  "nickname" = $1 or
  "email" = $2
;`
		_, err = conn.Prepare("UserCreateConflict", sql)
		return
	})
}

// особенность - err != nil при ошибке базы, при вставке нарушающих
// уникальность значений len(conflictUsers) != 0
func (cp *ConnPool) UserCreate(user types.User) (err error, conflictUsers types.Users) {
	_, err = cp.Exec("UserCreate", user.Nickname, user.About, user.Email, user.Fullname)
	if err != nil {
		if err.(pgx.PgError).Code == "23505" {
			// if "Duplicate key value violates unique constraint, key already exists."
			var rows *pgx.Rows
			rows, err = cp.Query("UserCreateConflict",  user.Nickname, user.Email)
			defer rows.Close()
			for rows.Next() {
				user := types.User{}
				err = rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
				if err != nil {
					return
				}
				conflictUsers = append(conflictUsers, user)
			}
			err = rows.Err()
		}
	}
	return
}
