package accessor

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"reflect"
	"runtime"
)

// Глобальный регистратор для функций подготовки,
// собирает все добавленные функции подготовк и в одну.
// непотокобезопасно, все init() функции собираются в один исполняемый поток.
type Preparer struct {
	functionsToPrepare []func(*pgx.Conn) (error)
}

func (p *Preparer) add(function func(*pgx.Conn)(error)) {
	p.functionsToPrepare = append(p.functionsToPrepare, function)
}

// Добавление привелегированной функции, которая точно будет вызвана первой.
// Используется для создания таблиц и идексов, до подготовки всяких 'select'.
// Вызывать 1 раз.
func (p *Preparer) addFirst(function func(*pgx.Conn)(error)) {
	if len(p.functionsToPrepare) == 0 {
		p.functionsToPrepare = append(p.functionsToPrepare, function)
	} else {
		p.functionsToPrepare = append([]func(*pgx.Conn)(error){function}, p.functionsToPrepare...) // ... - операция распаковки массива как **[] в python3.
	}
}

func (p *Preparer) Execute(conn *pgx.Conn) (err error) {
	for _, function := range p.functionsToPrepare {
		if err := function(conn); err != nil {
			return errors.New("error on execute function '" +
				runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name() +
				"' :" + err.Error())
		}
	}
	return nil
}

// Это статический объект, так же как и init функции, что добавляют функции подготовки сюда.
var Prep Preparer
