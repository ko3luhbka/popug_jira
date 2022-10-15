package rest

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ko3luhbka/task_tracker/rest/model"
)

func (s Server) ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON("pong")
}

func (s Server) createTask(c *fiber.Ctx) error {
	var t model.Task
	if err := c.BodyParser(&t); err != nil {
		log.Printf("failed to parse body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if err := t.ValidateCreate(); err != nil {
		log.Printf("invalid task: %v\n", err)
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	}

	created, err := s.Svc.CreateTask(c.Context(), t)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (s Server) getAllTasks(c *fiber.Ctx) error {
	tasks, err := s.Svc.GetAllTasks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(tasks)
}

func (s Server) getTask(c *fiber.Ctx) error {
	id, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	t, err := s.Svc.GetTaskByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(t)
}

func (s Server) updateTask(c *fiber.Ctx) error {
	var t model.Task
	uuid, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := c.BodyParser(&t); err != nil {
		log.Printf("failed to parse body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := t.ValidateUpdate(); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).SendString(err.Error())
	}
	t.ID = uuid

	updated, err := s.Svc.UpdateTask(c.Context(), t)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(updated)
}

func (s Server) deleteTask(c *fiber.Ctx) error {
	id, err := s.parseID(c)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := s.Svc.DeleteTask(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

func (s Server) reassignTasks(c *fiber.Ctx) error {
	if err := s.Svc.ReassignTasks(c.Context()); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

func (s Server) parseID(ctx *fiber.Ctx) (string, error) {
	idParam := ctx.Params("id")
	if idParam == "" {
		err := fmt.Errorf("task id is empty")
		log.Println(err)
		return "", err
	}
	return idParam, nil
}
