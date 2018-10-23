package types

import (
	"time"
)

// Ветка обсуждения на форуме.
//easyjson:json
type Thread struct {
	// Пользователь, создавший данную тему.
	Author string `json:"author"`
	// Дата создания ветки на форуме.
	Created time.Time `json:"created,omitempty"`
	// Форум, в котором расположена данная ветка обсуждения.
	Forum string `json:"forum,omitempty"`
	// Идентификатор ветки обсуждения.
	Id uint `json:"id,omitempty"`
	// Описание ветки обсуждения.
	Message string `json:"message"`
	// Человекопонятный URL (https://ru.wikipedia.org/wiki/Семантический_URL).
	// В данной структуре slug опционален и не может быть числом.
	Slug string `json:"slug,omitempty"`
	// Заголовок ветки обсуждения.
	Title string `json:"title"`
	// Кол-во голосов непосредственно за данное сообщение форума.
	Votes uint `json:"votes,omitempty"`
}
