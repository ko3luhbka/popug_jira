package rest

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ko3luhbka/task_tracker/db"
	"github.com/ko3luhbka/task_tracker/mq"
	"github.com/ko3luhbka/task_tracker/service"
)

const (
	listenAddr = "0.0.0.0:8081"
	baseURL    = "/"
)

type Server struct {
	Svc *service.Service
	app *fiber.App
}

func NewServer(tr *db.TaskRepo, ar *db.AssigneeRepo, mq *mq.Client) (*Server, error) {
	var appCfg = fiber.Config{
		CaseSensitive: true,
		StrictRouting: false,
	}

	app := fiber.New(appCfg)
	app.Use(logger.New())

	svc := service.NewService(tr, ar, mq)

	srv := &Server{
		Svc: svc,
		app: app,
	}

	srv.initRoutes()
	return srv, nil
}

func (s Server) Run(errCh chan<- error) {
	go func() {
		if err := s.app.Listen(listenAddr); err != nil {
			errCh <- err
		}
	}()
}

func (s Server) Shutdown() error {
	return s.app.Shutdown()
}

func (s Server) initRoutes() {
	base := s.app.Group(baseURL)
	base.Get("/ping", s.ping)

	tasks := base.Group("tasks")
	tasks.Post("/", s.createTask)
	tasks.Get("/", s.getAllTasks)
	tasks.Get("/:id", s.getTask)
	tasks.Patch("/:id", s.updateTask)
	tasks.Delete("/:id", s.deleteTask)
	tasks.Post("/reassign", s.reassignTasks)
}

func parseBody(c *fiber.Ctx, object any) error {
	if err := c.BodyParser(object); err != nil {
		log.Printf("request body parsing error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON("invalid request body")
	}
	return nil
}
