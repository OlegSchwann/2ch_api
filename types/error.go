package types

//easyjson:json
type Error struct {
	// Текстовое описание ошибки.
	// В процессе проверки API никаких проверок на содерижимое данного описания не делается.
	Message string `json:"message,omitempty"`
}
