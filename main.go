// Тестовое задание для реализации проекта \"Форумы\" на курсе по базам данных в Технопарке Mail.ru (https://park.mail.ru).

package main

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/log15adapter"
	"github.com/valyala/fasthttp"
	log "gopkg.in/inconshreveable/log15.v2"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/global_environment"
	"github.com/OlegSchwann/2ch_api/router"
)

func main() {
	// Создаём хранилище глобальных переменных.
	env := global_environment.Environment{}

	if false {
		// Инициализируем логгер, результаты выводятся в stdout.
		env.Logger = log15adapter.NewLogger(log.New("module", "pgx"))
	} else {
		// передаём пустышку.
		env.Logger = &LoggerStub{}
	}

	// Вытаскиваем статический объект пакета,
	// агрегатор SQL выражений, которые надо подготовить в базе данных.
	env.Prep = &accessor.Prep

	// Парсим конфиг из файла config.json, что лежит рядом с бинарником,
	// в репозитории это 'github.com/OlegSchwann/2ch_api/_build_configs/config.json'
	if len(os.Args) < 2 {
		log.Crit("you must pass the path to the configuration file as the first argument")
		os.Exit(1)
	}
	configBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Crit("unable to open configuration file '"+os.Args[1]+"': "+err.Error())
		os.Exit(1)
	}
	err = env.Config.UnmarshalJSON(configBytes)
	if err != nil {
		log.Crit("unable to parse configuration file ./config.json : "+
			"it should be json with all fields : "+err.Error())
		os.Exit(1)
	}

	// Устанавливаем соединение с базой данных.
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     env.Config.DatabaseHost,
			Port:     env.Config.DatabasePort,
			User:     env.Config.DatabaseUser,
			Password: env.Config.DatabasePassword,
			Database: env.Config.DatabaseSpace,
			Logger:   env.Logger,
		},
		MaxConnections: 8, // именно во столько будет проводиться нагрузочное тестирование.
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

	// Регистируем обработчики Url.
	env.Router = router.RegisterHandlers(&env)

	// Запускаем сервер.
	serverPort := ":" + strconv.Itoa(int(env.Config.ServerPort))
	log.Debug("Server started on http://[::1]:"+serverPort+"/")
	log.Error(fasthttp.ListenAndServe(serverPort, env.Router.Handler).Error())
}

type LoggerStub struct {}
func (LoggerStub) Log(level pgx.LogLevel, msg string, data map[string]interface{}){
	// Выйти, не тратя процессорное время на запись ненужных строк.
}
