package handlers

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
)

func (e *Environment)ThreadGetPosts(ctx *fasthttp.RequestCtx) {
	slugOrId := ctx.UserValue("slug_or_id").(string) // Идентификатор ветки обсуждения.

	args := ctx.URI().QueryArgs()
	limit, err := args.GetUint("limit") // Максимальное кол-во возвращаемых записей.
	if err != nil {
		limit = 100
	}

	since, err := args.GetUint("since") // Идентификатор поста, после которого будут выводиться записи
	if err != nil { //                         пост с данным идентификатором в результат не попадает).

	}

	switch string(args.Peek("sort")) {
	case "flat": // по дате, комментарии выводятся простым списком в порядке создания;
	case "tree": // древовидный, комментарии выводятся отсортированные в дереве по N штук;
	case "parent_tree": // древовидные с пагинацией по родительским (parent_tree), на странице N родительских комментов и все комментарии прикрепленные к ним, в древвидном отображение. Подробности: https://park.mail.ru/blog/topic/view/1191/
	case "": // По умолчанию: flat
	}

	desc := args.GetBool("desc")


	fmt.Print(slugOrId, limit, since, desc)


	ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")
	ctx.Response.Header.SetStatusCode(http.StatusOK)
}

