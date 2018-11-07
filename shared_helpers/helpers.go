package shared_helpers

import (
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

func ZeroPad(integer uint, overallLen int) string {
	num := strconv.FormatUint(uint64(integer), 10)
	return strings.Repeat("0", overallLen-len(num)) + num
}

// объявляет функцию, как возврающую json
func ContentTypeJson(wrapped func(ctx *fasthttp.RequestCtx) ()) (func(ctx *fasthttp.RequestCtx) ()) {
	return func(ctx *fasthttp.RequestCtx) () {
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
		wrapped(ctx)
	}
}

// TODO: зарефакторить обработку ошибок.
//// существенно уменьшает объём кода в handlers - аварийно завершает выполнение,
//// переходит к shared_helpers.Recover()
//func MustBeNil(err error, additionalMessageIfErr string){
//	if err != nil {
//		panic(additionalMessageIfErr + ": " + err.Error())
//	}
//	return
//}
//
//// ловит панику (паникуем на ошибки, которые не знаем, как обработать(пока)),
//// записывает ошибку в формате type.Error, устанавливает код ответа "internal server error".
//func Recover(wrapped func(ctx *fasthttp.RequestCtx) ()) (func(ctx *fasthttp.RequestCtx) ()) {
//	return func(ctx *fasthttp.RequestCtx) () {
//		defer func() {
//			if err := recover(); err != nil {
//				response, _ := types.Error{
//					Message: fmt.Sprintf("%#v", err),
//				}.MarshalJSON()
//				ctx.Write(response)
//				ctx.Response.Header.SetStatusCode(http.StatusInternalServerError)
//			}
//		}()
//		wrapped(ctx)
//	}
//}
