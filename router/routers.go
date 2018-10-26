package router

import (
	"github.com/OlegSchwann/2ch_api/global_environment"
	"github.com/OlegSchwann/2ch_api/handlers"
	"github.com/buaazp/fasthttprouter"
)

func RegisterHandlers(environment *global_environment.Environment) (router *fasthttprouter.Router) {
	env := handlers.Environment{Environment: *environment}
	router = fasthttprouter.New()
	router.GET("/", env.HealthCheck)                                   //
	router.POST("/api/service/clear", env.ServiceClear)                //
	router.GET("/api/service/status", env.ServiceStatus)               //
	router.POST("/api/forum/:slug" /*only "create"*/, env.ForumCreate) //
	router.GET("/api/forum/:slug/details", env.ForumGetDetails)
	router.GET("/api/forum/:slug/threads", env.ForumGetThreads)
	router.GET("/api/forum/:slug/users", env.ForumGetUsers)
	router.POST("/api/forum/:slug/create", env.ThreadCreate) //
	router.POST("/api/thread/:slug_or_id/create", env.PostsCreate)
	router.GET("/api/thread/:slug_or_id/details", env.ThreadGetDetails) //
	router.POST("/api/thread/:slug_or_id/details", env.ThreadUpdateDetails)
	router.GET("/api/thread/:slug_or_id/posts", env.ThreadGetPosts)
	router.POST("/api/thread/:slug_or_id/vote", env.ThreadVote)
	router.GET("/api/post/:id/details", env.PostGetDetails)
	router.POST("/api/post/:id/details", env.PostUpdateDetails)
	router.POST("/api/user/:nickname/create", env.UserCreate)         //
	router.GET("/api/user/:nickname/profile", env.UserGetProfile)     //
	router.POST("/api/user/:nickname/profile", env.UserUpdateProfile) //
	return
}
