package main

import (
	"fmt"
	"log"
	"park_db_course/cfg"
	httphandlers "park_db_course/internal/api/http"
	"park_db_course/internal/repository"

	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()

	dsn := fmt.Sprintf(`user=%s dbname=%s password=%s host=%s port=%s sslmode=disable`,
		cfg.DBUser, cfg.DBName, cfg.DBPassword, cfg.DBHost, cfg.DBPort)

	conn, err := pgx.ParseConnectionString(dsn)
	if err != nil {
		log.Fatalln("cant parse cfg", err)
	}

	db, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     conn,
		MaxConnections: cfg.DBMaxCon,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	})

	if err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepo(db)
	forumRepo := repository.NewForumRepo(db)
	threadRepo := repository.NewThreadRepo(db)
	postRepo := repository.NewPostRepo(db)
	serviceRepo := repository.NewServiceRepo(db)

	userH := httphandlers.NewUserH(userRepo)
	forumH := httphandlers.NewForumH(forumRepo, userRepo, threadRepo)
	threadH := httphandlers.NewThreadH(threadRepo, userRepo)
	postH := httphandlers.NewPostH(postRepo)
	serviceH := httphandlers.NewServiceH(serviceRepo)

	// Register routes
	// ---------------
	// forum
	r.POST("/api/forum/create", forumH.Create)
	r.GET("/api/forum/{slug}/details", forumH.Details)
	r.POST("/api/forum/{slug}/create", forumH.CreateThread)
	r.GET("/api/forum/{slug}/threads", forumH.ForumThreads)
	r.GET("/api/forum/{slug}/users", forumH.ForumUsers)
	// post
	r.GET("/api/post/{id}/details", postH.GetDetails)
	r.POST("/api/post/{id}/details", postH.UpdateDetails)
	// service
	r.GET("/api/service/status", serviceH.Status)
	r.POST("/api/service/clear", serviceH.Clear)
	// thread
	r.POST("/api/thread/{slug_or_id}/create", threadH.CreatePost)
	r.POST("/api/thread/{slug_or_id}/vote", threadH.CreateVote)
	r.GET("/api/thread/{slug_or_id}/details", threadH.Details)
	r.GET("/api/thread/{slug_or_id}/posts", threadH.ThreadPost)
	r.POST("/api/thread/{slug_or_id}/details", threadH.Update)
	// user
	r.POST("/api/user/{nickname}/create", userH.Create)
	r.GET("/api/user/{nickname}/profile", userH.GetByNickname)
	r.POST("/api/user/{nickname}/profile", userH.Update)

	fmt.Println("[SERVICE STARTED]", cfg.ApiPort)

	log.Fatal(fasthttp.ListenAndServe(cfg.ApiPort, r.Handler))
}
