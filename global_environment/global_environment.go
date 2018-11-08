package global_environment

import (
	"github.com/OlegSchwann/2ch_api/types"
	"github.com/buaazp/fasthttprouter"
	log "github.com/jackc/pgx/log/log15adapter"

	"github.com/OlegSchwann/2ch_api/accessor"
)

type Environment struct {
	Logger   *log.Logger
	ConnPool *accessor.ConnPool
	Prep     *accessor.Preparer
	Config   types.Config
	Router   *fasthttprouter.Router
}
