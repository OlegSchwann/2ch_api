package types

// Информация о пользователе.
//easyjson:json
type UserUpdate struct {
	// Имя пользователя. Приходит в пути, а не в json.
	Nickname string `json:"-"`
	// Описание пользователя.
	About string `json:"about,omitempty"`
	// Почтовый адрес пользователя (уникальное поле).
	Email string `json:"email,omitempty"`
	// Полное имя пользователя.
	Fullname string `json:"fullname,omitempty"`
}
