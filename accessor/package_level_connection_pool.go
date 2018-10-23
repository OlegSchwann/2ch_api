package accessor

import "github.com/jackc/pgx"

// глобальный объект пакета с пулом соединений.
// Вся SQL логика прикрепляется к нему как методы.
type ConnPool struct {
	pgx.ConnPool
}
