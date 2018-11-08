package types

// Конфигурация подключения к базе и запуска сервера.

//easyjson:json
type Config struct {                                           // {
	DatabaseHost     string `json:"DatabaseHost,required"`     // "DatabaseHost":     "127.0.0.1",
	DatabasePort     uint16 `json:"DatabasePort,required"`     // "DatabasePort":     5432,
	DatabaseUser     string `json:"DatabaseUser,required"`     // "DatabaseUser":     "postgres",
	DatabasePassword string `json:"DatabasePassword,required"` // "DatabasePassword": "",
	DatabaseSpace    string `json:"DatabaseSpace,required"`    // "DatabaseSpace":    "postgres",
	ServerPort       uint16 `json:"ServerPort,required"`       // "ServerPort":       5000
}                                                              // }
