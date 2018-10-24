package accessor

import (
	"github.com/jackc/pgx"

	"github.com/OlegSchwann/2ch_api/types"
)

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  (select count(*) from "forum" ) as "count_forum",
  (select count(*) from "post"  ) as "count_post",
  (select count(*) from "thread") as "count_thread",
  (select count(*) from "user"  ) as "count_user";`
		_, err = conn.Prepare("ServiceStatus", sql)
		return err
	})
}

func (cp *ConnPool) ServiceStatus() (status types.Status, err error) {
	err = cp.QueryRow("ServiceStatus").Scan(
		&status.Forum, &status.Post, &status.Thread, &status.User)
	return
}
