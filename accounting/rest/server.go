package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/ko3luhbka/accounting/db"
	"github.com/ko3luhbka/accounting/mq"
	"github.com/ko3luhbka/accounting/service"
)

const (
	listenAddr = "0.0.0.0:8082"
	baseURL    = "/"
)

type Server struct {
	Svc *service.Service
	app *fiber.App
}

func NewServer(tr *db.AccountRepo, ar *db.AuditRepo, mq *mq.Client) (*Server, error) {
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

	base.Get("/user/:id/audit", s.getUserAuditLog)
	base.Get("/user/:id/balance", s.getUserBalance)
	base.Get("/management-income", s.getManagementIncome)
}
