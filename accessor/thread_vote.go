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
  "id"
from
  "thread"
where 
  slug = $1
;`
		_, err = conn.Prepare("ThreadVoteGetThreadIdBySlug", sql)
		return
	})
}

// Превращаем "thread"."slug" в "thread"."id" .
func (cp *ConnPool) ThreadVoteGetThreadIdBySlug(threadSlug string) (threadId int, err error) {
	err = cp.QueryRow("ThreadVoteGetThreadIdBySlug", threadSlug).Scan(&threadId)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}

// запрашиваем, существует ли такая запись в таблице, и с каким голосом.
//                                       ┌──true──<запись существует?>──false──┐
//  ┌──true──<меняется ли голос на противоположный?>──false──┐       ⎡insert vote;       ⎤
//⎡update vote;           ⎤                [запрашиваем thread;]     ⎢thread.vote += vote⎥
//⎢thread.vote += 2 * vote⎥                                          ⎣возвращая thread;  ⎦
//⎣возвращая thread;      ⎦

func init() {
	// TODO: оптимизировать?
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select 
  "voice"
from
  "vote"
where 
  "nickname" = $1 and 
  "thread_id" = $2
;`
		_, err = conn.Prepare("ThreadVoteSelectOldVote", sql)
		return
	})
}

func (cp *ConnPool) ThreadVoteSelectOldVote(nickname string, threadId int) (oldVote int8, err error) {
	err = cp.QueryRow("ThreadVoteSelectOldVote", nickname, threadId).Scan(&oldVote)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
insert into "vote" (
  "nickname",
  "voice",
  "thread_id"
) values (
  $1,
  $2,
  $3
);`
		_, err = conn.Prepare("ThreadVoteInsert", sql)
		return
	})
}

// Вставляем новый голос в таблицу голосов
func (cp *ConnPool) ThreadVoteInsert(nickname string, voice int8, id int) (err error) {
	_, err = cp.Exec("ThreadVoteInsert", nickname, voice, id)
	if err != nil {
		pgxPgError := err.(pgx.PgError)
		if pgxPgError.Code == "23503" {
			// Insert or update on table "post" violates foreign key constraint.
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		} else if pgxPgError.Code == "23505" {
			// Duplicate key value violates unique constraint "vote_nickname_id_unique".
			err = &Error{
				Code:            http.StatusConflict,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}

// если не удалось предыдущее обновляем голос в таблице.
// Существует конструкция insert ... on conflict update set ... ;
// но она не позволяет выяснить, что произошло: вставка или обновление.

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update "vote" set 
  "voice" = $1
where 
  "nickname" = $2 and
  "thread_id" = $3
;`
		_, err = conn.Prepare("ThreadVoteUpdateVote", sql)
		return
	})
}

// Обновляем голос в таблице голосов.
func (cp *ConnPool) ThreadVoteUpdateVote(nickname string, voice int8, threadId int) (err error) {
	_, err = cp.Exec("ThreadVoteUpdateVote", voice, nickname, threadId)
	if err != nil {
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}

// Oбновляем денормализованное значение суммы голосов в строке конкретного треда.
func init() {
	// TODO: оптимизировать.
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
update
  "thread"
set 
  "votes" = "votes" + $1
where 
  "id" = $2
returning 
  "author",
  "created",
  "forum",
  "id",
  "message",
  "slug",
  "title",
  "votes"
;`
		_, err = conn.Prepare("ThreadVoteUpdateThreadVote", sql)
		return
	})
}

func (cp *ConnPool) ThreadVoteUpdateThreadVote(voiceDelta int8, threadId int) (thread types.Thread, err error) {
	err = cp.QueryRow("ThreadVoteUpdateThreadVote", voiceDelta, threadId).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}

func init() {
	Prep.add(func(conn *pgx.Conn) (err error) {
		// language=PostgreSQL
		sql := `
select
  "author",
  "created",
  "forum",
  "id",
  "message",
  "slug",
  "title",
  "votes"
from
  "thread"
where 
  "id" = $1
;`
		_, err = conn.Prepare("ThreadVoteSelectUnchangedThread", sql)
		return
	})
}

// возвращаем информацию о thread, с суммой голосов.
func (cp *ConnPool) ThreadVoteSelectUnchangedThread(threadId int) (thread types.Thread, err error) {
	err = cp.QueryRow("ThreadVoteSelectUnchangedThread", threadId).Scan(
		&thread.Author, &thread.Created, &thread.Forum, &thread.Id, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		if err.Error() == "no rows in result set" {
			err = &Error{
				Code:            http.StatusNotFound,
				UnderlyingError: err,
			}
			return
		}
		err = &Error{
			Code:            http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	return
}
