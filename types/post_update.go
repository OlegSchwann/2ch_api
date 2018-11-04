package types

// Сообщение для обновления сообщения внутри ветки на форуме.
// Пустые параметры остаются без изменений.
//easyjson:json
type PostUpdate struct {
	// Собственно сообщение форума.
	Message *string `json:"message,omitempty"`
}
