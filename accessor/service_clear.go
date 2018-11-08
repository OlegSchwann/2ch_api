package accessor

func (cp *ConnPool)ServiceClear() (err error) {
	// language=PostgreSQL
	sql := `
truncate table only 
  "user",
  "forum",
  "thread",
  "vote",
  "post",
  "user_in_forum"
restart identity restrict;`
	_, err = cp.Exec(sql)
	return
}
