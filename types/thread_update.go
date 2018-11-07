package types

// Сообщение для обновления ветки обсуждения на форуме.
// Пустые параметры остаются без изменений.
//easyjson:json
type ThreadUpdate struct {
	// Описание ветки обсуждения.
	Message *string `json:"message,omitempty"`
	// Заголовок ветки обсуждения.
	Title *string `json:"title,omitempty"`
}
