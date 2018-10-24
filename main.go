// Тестовое задание для реализации проекта \"Форумы\" на курсе по базам данных в Технопарке Mail.ru (https://park.mail.ru).

package main

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/log15adapter"
	"github.com/valyala/fasthttp"
	log "gopkg.in/inconshreveable/log15.v2"
	"os"
	"strconv"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/global_environment"
	"github.com/OlegSchwann/2ch_api/router"
)

func main() {
	// создаём хранилище глобальных переменных
	env := global_environment.Environment{}
	// инициализируем логгер
	env.Logger = log15adapter.NewLogger(log.New("module", "pgx"))
	// вытаскиваем статический объект из пакета.
	// Это фасад для выражений, которые надо подготовить в базе данных.
	env.Prep = &accessor.Prep
	// парсим конфиг
	env.Config = map[string]string{
		// TODO: парсить из файла и проверять наличие необходимых полей.
		"host":     "127.0.0.1",
		"port":     "5432",
		"user":     "postgres",
		"password": "",
		"database": "postgres",
	}
	// устанавливаем соединение с базой данных
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host: env.Config["host"],
			Port: func() (port uint16) {
				if portString, ok := env.Config["port"]; ok {
					port, err := strconv.Atoi(portString)
					if err != nil || port > 65535 || port < 1 {
						panic("port must be 1-65535, default 5432, got '" + portString + "'.")
					}
				} else {
					port = 5432
				}
				return
			}(),
			User:     env.Config["postgres"],
			Password: env.Config["password"],
			Database: env.Config["database"],
			Logger:   env.Logger,
		},
		MaxConnections: 5,
		// Создаём таблицы в базе данных,
		// Компилируем sql запросы для каждого соединения после их установления.
		AfterConnect: env.Prep.Execute,
	})
	if err != nil {
		log.Crit("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	env.ConnPool = &accessor.ConnPool{ConnPool: *pool}
	defer env.ConnPool.Close()
	// регистируем обработчики Url
	env.Router = router.RegisterHandlers(&env)
	env.Logger.Log(pgx.LogLevelInfo, "Server started on http://127.0.0.1:8080/", map[string]interface{}{})

	// закончена инициализация global_environment.
	// cлушаем на порту 8080.
	log.Error(fasthttp.ListenAndServe(":8080", env.Router.Handler).Error()) // TODO: graceful shutdown
}
