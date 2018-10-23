package types

// Информация о пользователе.
//easyjson:json
type UserUpdate struct {
	// Описание пользователя.
	About string `json:"about,omitempty"`
	// Почтовый адрес пользователя (уникальное поле).
	Email string `json:"email,omitempty"`
	// Полное имя пользователя.
	Fullname string `json:"fullname,omitempty"`
}
