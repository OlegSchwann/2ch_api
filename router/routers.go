package router

import (
	"github.com/OlegSchwann/2ch_api/global_environment"
	"github.com/OlegSchwann/2ch_api/handlers"
	"github.com/OlegSchwann/2ch_api/shared_helpers"
	"github.com/buaazp/fasthttprouter"
)

// Middleware
var m1 = shared_helpers.ContentTypeJson
// var m2 = shared_helpers.Recover

func RegisterHandlers(environment *global_environment.Environment) (router *fasthttprouter.Router) {
	e := handlers.Environment{Environment: *environment}
	router = fasthttprouter.New()
	router.GET ("/",                                m1(e.HealthCheck        )) //
	router.POST("/api/service/clear",               m1(e.ServiceClear       )) //
	router.GET ("/api/service/status",              m1(e.ServiceStatus      )) //
	router.POST("/api/forum/:slug"/*only "create"*/,m1(e.ForumCreate        )) //
	router.GET ("/api/forum/:slug/details",         m1(e.ForumGetDetails    )) //
	router.GET ("/api/forum/:slug/threads",         m1(e.ForumGetThreads    )) //
	router.GET ("/api/forum/:slug/users",           m1(e.ForumGetUsers      ))
	router.POST("/api/forum/:slug/create",          m1(e.ThreadCreate       )) //
	router.POST("/api/thread/:slug_or_id/create",   m1(e.PostsCreate        )) //
	router.GET ("/api/thread/:slug_or_id/details",  m1(e.ThreadGetDetails   )) //
	router.POST("/api/thread/:slug_or_id/details",  m1(e.ThreadUpdateDetails)) //
	router.GET ("/api/thread/:slug_or_id/posts",    m1(e.ThreadGetPosts     )) //
	router.POST("/api/thread/:slug_or_id/vote",     m1(e.ThreadVote         )) //
	router.GET ("/api/post/:id/details",            m1(e.PostGetDetails     )) //
	router.POST("/api/post/:id/details",            m1(e.PostUpdateDetails  )) //
	router.POST("/api/user/:nickname/create",       m1(e.UserCreate         )) //
	router.GET ("/api/user/:nickname/profile",      m1(e.UserGetProfile     )) //
	router.POST("/api/user/:nickname/profile",      m1(e.UserUpdateProfile  )) //
	return
}
