package types

// Информация о форуме.
//easyjson:json
type Forum struct {
	// Общее кол-во сообщений в данном форуме.
	Posts   uint `json:"posts,omitempty"`
	// Человекопонятный URL (https://ru.wikipedia.org/wiki/Семантический_URL), уникальное поле.
	Slug    string  `json:"slug"`
	// Общее кол-во ветвей обсуждения в данном форуме. 
	Threads uint `json:"threads,omitempty"`
	// Название форума.
	Title   string  `json:"title"`
	// Nickname пользователя, который отвечает за форум.
	User    string  `json:"user"`
}
