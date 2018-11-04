package types

import (
	"time"
)

// Сообщение внутри ветки обсуждения на форуме.
//easyjson:json
type Post struct {
	// Автор, написавший данное сообщение.
	Author           string    `json:"author"`
	// Дата создания сообщения на форуме.
	Created          time.Time `json:"created,omitempty"`
	// Идентификатор форума (slug) данного сообещния.
	Forum            string    `json:"forum,omitempty"`
	// Идентификатор данного сообщения.
	Id               int       `json:"id,omitempty"`
	// Истина, если данное сообщение было изменено.
	IsEdited         bool      `json:"isEdited,omitempty"`
	// Собственно сообщение форума.
	Message          string    `json:"message"`
	// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Parent           int       `json:"parent,omitempty"`
	// Идентификатор ветви (id) обсуждения данного сообещния.
	ThreadId         int       `json:"thread,omitempty"`
	// Идентификатор ветки (slug) обсуждения, опционально. 
	ThreadSlug       string    `json:"-"`
	// Материализованный путь в дереве сообщений
	MaterializedPath string    `json:"-"`
	// количество дочерних постов.
	NumberOfChildren int       `json:"-"`
}
