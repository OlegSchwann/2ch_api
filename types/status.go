package types

//easyjson:json
type Status struct {
	// Количество разделов в базе данных.
	Forum uint `json:"forum"`
	// Количество сообщений в базе данных.
	Post uint `json:"post"`
	// Количество веток обсуждения в базе данных.
	Thread uint `json:"thread"`
	// Количество пользователей в базе данных.
	User uint `json:"user"`
}
