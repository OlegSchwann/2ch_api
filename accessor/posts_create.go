package accessor

import (
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// QuoteString escapes and quotes a string making it safe for interpolation into an SQL string.
func quoteString(input string) (output string) {
	output = "'" + strings.Replace(input, "'", "''", -1) + "'"
	return
}

func (cp *ConnPool) PostsCreate(posts types.Posts, treadSlug string) (responsePosts types.Posts, err error) {
	buffer := strings.Builder{}
	buffer.Write([]byte(`
insert into "post" (
  "thread_id",
  "thread_slug",
  "author",
  "created",
  "message",
  "parent"
) values `))
	for i, post := range posts{
		if i == 0 {
			buffer.Write([]byte("("))
		} else {
			buffer.Write([]byte(", ("))
		}
		/* thread_id   */buffer.Write([]byte(strconv.Itoa(int(post.Thread))));    buffer.Write([]byte(", "))
		/* thread_slug */buffer.Write([]byte(quoteString(treadSlug)));            buffer.Write([]byte(", "))
		/* author      */buffer.Write([]byte(quoteString(post.Author)));          buffer.Write([]byte(", '"))
		/* created     */buffer.Write([]byte(post.Created.Format(time.RFC3339))); buffer.Write([]byte("', "))
		/* message     */buffer.Write([]byte(quoteString(post.Message)));         buffer.Write([]byte(", "))
		/* parent      */buffer.Write([]byte(strconv.Itoa(int(post.Parent))));    buffer.Write([]byte(")"))
	}
	buffer.Write([]byte(`
returning 
"author",
"created",
"forum",
"id",            
"message",
"parent",
"thread_id";`))

	rows, err := cp.Query(buffer.String())
	if err != nil {
		err = &Error{
			Code: http.StatusInternalServerError,
			UnderlyingError: err,
		}
		return
	}
	defer rows.Close()

	for rows.Next() {
		post := types.Post{}
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			err = &Error{
				Code: http.StatusInternalServerError,
				UnderlyingError: err,
			}
			return
		}
		responsePosts = append(responsePosts, post)
	}
	// тут следовало бы поймать ошибку триггера, возвещающую о нарушении внешних ключей, но почему-то она не пробрасывается в результаты запроса.
	// просто проверяем, что бы всё вставилось.
	if len(posts) != len(responsePosts) {
		err = &Error{
			Code: http.StatusNotFound,
			UnderlyingError: errors.New("The trigger worked with an error; not so many fields were inserted as was required."),
		}
	}
	return
}
