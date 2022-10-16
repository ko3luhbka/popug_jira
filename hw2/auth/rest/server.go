package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/ko3luhbka/auth/db"
	"github.com/ko3luhbka/auth/mq"
)

const (
	listenAddr = "0.0.0.0:8080"
	baseURL    = "/"
)

type Server struct {
	mq   *mq.Client
	repo *db.Repo
	app  *fiber.App
}

func NewServer(repo *db.Repo, mq *mq.Client) (*Server, error) {
	var appCfg = fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
	}

	app := fiber.New(appCfg)
	app.Use(logger.New())

	srv := &Server{
		repo: repo,
		app:  app,
		mq:   mq,
	}

	srv.initRoutes()
	return srv, nil
}

func (s Server) Run() error {
	if err := s.app.Listen(listenAddr); err != nil {
		return err
	}
	return nil
}

func (s Server) Shutdown() error {
	return s.app.Shutdown()
}

func (s Server) initRoutes() {
	base := s.app.Group(baseURL)
	base.Get("/ping", s.ping)

	base.Post("/users", s.createUser)
	base.Get("/users", s.getAllUsers)
	base.Get("/users/:id", s.getUser)
	base.Patch("/users/:id", s.updateUser)
	base.Delete("/users/:id", s.deleteUser)
}
