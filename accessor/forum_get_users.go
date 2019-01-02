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
  "user_in_forum"."nickname",
  "user"."about",            
  "user"."email",            
  "user"."fullname"          
from
  "user_in_forum",
  "user"
where
  "user_in_forum"."forum" = $1 and
  "user_in_forum"."nickname" = "user"."nickname"
order by
  "user_in_forum"."nickname" asc
limit $2
;`
		_, err = conn.Prepare("ForumGetUsersAsc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "user_in_forum"."nickname",
  "user"."about",            
  "user"."email",            
  "user"."fullname"          
from
  "user_in_forum",
  "user"
where
  "user_in_forum"."forum" = $1 and
  "user_in_forum"."nickname" = "user"."nickname"
order by
  "user_in_forum"."nickname" desc
limit $2
;`
		_, err = conn.Prepare("ForumGetUsersDesc", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "user_in_forum"."nickname",
  "user"."about",            
  "user"."email",            
  "user"."fullname"          
from
  "user_in_forum",
  "user"
where
  "user_in_forum"."forum" = $1 and
  "user_in_forum"."nickname" > $2 and
  "user_in_forum"."nickname" = "user"."nickname"
order by
  "user_in_forum"."nickname" asc
limit $3
;`
		_, err = conn.Prepare("ForumGetUsersAscSince", sql)
		return err
	})
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "user_in_forum"."nickname",
  "user"."about",            
  "user"."email",            
  "user"."fullname"          
from
  "user_in_forum",
  "user"
where
  "user_in_forum"."forum" = $1 and
  "user_in_forum"."nickname" < $2 and
  "user_in_forum"."nickname" = "user"."nickname"
order by
  "user_in_forum"."nickname" desc
limit $3
;`
		_, err = conn.Prepare("ForumGetUsersDescSince", sql)
		return err
	})
}

func (cp *ConnPool) ForumGetUsers(forumSlug string, limit int, desc bool) (users types.Users, err error) {
	var rows *pgx.Rows
	if desc {
		rows, err = cp.Query("ForumGetUsersDesc", forumSlug, limit)
	} else {
		rows, err = cp.Query("ForumGetUsersAsc", forumSlug, limit)
	}
	defer rows.Close()
	if err != nil {
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	for rows.Next() {
		user := types.User{}
		err = rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			err = &Error{
				Code: http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		users = append(users, user)
	}
	return
}

func (cp *ConnPool) ForumGetUsersSince(forumSlug string, limit int, sinceNickname string, desc bool) (users types.Users, err error) {
	var rows *pgx.Rows
	if desc {
	  	rows, err = cp.Query("ForumGetUsersDescSince", forumSlug, sinceNickname, limit)
	} else {
	  	rows, err = cp.Query("ForumGetUsersAscSince", forumSlug, sinceNickname, limit)
	}
	defer rows.Close()
	if err != nil {
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	for rows.Next() {
        user := types.User{}
		err = rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			err = &Error{
				Code: http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		users = append(users, user)
	}
	return
}
