package global_environment

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/jackc/pgx"

	"github.com/OlegSchwann/2ch_api/accessor"
	"github.com/OlegSchwann/2ch_api/types"
)

type Environment struct {
	Logger   pgx.Logger
	ConnPool *accessor.ConnPool
	Prep     *accessor.Preparer
	Config   types.Config
	Router   *fasthttprouter.Router
}
