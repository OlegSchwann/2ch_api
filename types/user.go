package types

// Информация о пользователе.
//easyjson:json
type User struct {
	// Имя пользователя (уникальное поле). Данное поле допускает только латиницу, цифры и знак подчеркивания. Сравнение имени регистронезависимо.
	Nickname string `json:"nickname,omitempty"`
	// Описание пользователя.
	About string `json:"about,omitempty"`
	// Почтовый адрес пользователя (уникальное поле).
	Email string `json:"email"`
	// Полное имя пользователя.
	Fullname string `json:"fullname"`
}
