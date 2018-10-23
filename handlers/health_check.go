package handlers

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"os"
	"strconv"
)

func (e *Environment)HealthCheck(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, ""+
		"Golang сервер работает на машине '"+
		func() (name string) { name, _ = os.Hostname(); return }()+
		" под номером процесса "+
		strconv.Itoa(os.Getegid())+".")
}
