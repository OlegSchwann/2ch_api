package global_environment

import (
	"2ch_api/accessor"
	"github.com/buaazp/fasthttprouter"
	log "github.com/jackc/pgx/log/log15adapter"
)

type Environment struct {
	Logger   *log.Logger
	ConnPool *accessor.ConnPool
	Prep     *accessor.Preparer
	Config   map[string]string
	Router   *fasthttprouter.Router
}
