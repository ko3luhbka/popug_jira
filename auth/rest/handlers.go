package rest

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ko3luhbka/auth/mq"
	"github.com/ko3luhbka/auth/rest/model"
)

func (s Server) ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("pong")
}

func (s Server) createUser(c *fiber.Ctx) error {
	var u model.User
	if err := c.BodyParser(&u); err != nil {
		log.Printf("failed to parse body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := u.Validate(); err != nil {
		log.Printf("invalid user: %v\n", err)
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	}

	created, err := s.repo.Create(c.Context(), *u.ToEntity())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	e := &mq.UserEvent{
		Name: mq.UserCreatedEvent,
		Data: model.EntityToAssignee(created),
	}
	if err := s.mq.Produce(c.Context(), e); err != nil {
		log.Println(err)
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (s Server) getAllUsers(c *fiber.Ctx) error {
	users, err := s.repo.GetAll(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(users)
}

func (s Server) getUser(c *fiber.Ctx) error {
	id, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	u, err := s.repo.GetByID(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(u)
}

func (s Server) updateUser(c *fiber.Ctx) error {
	var u model.User
	uuid, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := c.BodyParser(&u); err != nil {
		log.Printf("failed to parse body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	u.ID = uuid

	updated, err := s.repo.Update(c.Context(), *u.ToEntity())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	e := &mq.UserEvent{
		Name: mq.UserUpdatedEvent,
		Data: model.EntityToAssignee(updated),
	}
	if err := s.mq.Produce(c.Context(), e); err != nil {
		log.Println(err)
	}
	return c.Status(fiber.StatusOK).JSON(updated)
}

func (s Server) deleteUser(c *fiber.Ctx) error {
	id, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := s.repo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	e := &mq.UserEvent{
		Name: mq.UserDeletedEvent,
		Data: &model.Assignee{
			ID: id,
		},
	}
	if err := s.mq.Produce(c.Context(), e); err != nil {
		log.Println(err)
	}

	return c.SendStatus(fiber.StatusOK)
}

// func (s Server) authUser(c *fiber.Ctx) error {
// TODO
// }

func (s Server) parseID(ctx *fiber.Ctx) (string, error) {
	idParam := ctx.Params("id")
	if idParam == "" {
		err := fmt.Errorf("user id is empty")
		log.Println(err)
		return "", err
	}
	return idParam, nil
}
