package rest

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (s Server) ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("pong")
}

func (s Server) getUserBalance(c *fiber.Ctx) error {
	userUUID, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	balance, err := s.Svc.GetUserBalance(c.Context(), userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).SendString(strconv.Itoa(balance))
}

func (s Server) getUserAuditLog(c *fiber.Ctx) error {
	userUUID, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	auditLog, err := s.Svc.GetUserAuditLog(c.Context(), userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(auditLog)
}

func (s Server) getManagementIncome(c *fiber.Ctx) error {
	income, err := s.Svc.GetManagementIncome(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).SendString(strconv.Itoa(income))
}

func (s Server) parseID(ctx *fiber.Ctx) (string, error) {
	idParam := ctx.Params("id")
	if idParam == "" {
		err := fmt.Errorf("id param is empty")
		log.Println(err)
		return "", err
	}
	return idParam, nil
}
