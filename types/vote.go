package types

// Информация о голосовании пользователя.
//easyjson:json
type Vote struct {
	// Идентификатор пользователя.
	Nickname string `json:"nickname"`
	// Отданный голос ∈ [-1, 1].
	Voice int8 `json:"voice"`
}
