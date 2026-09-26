package api

import "github.com/gofiber/fiber/v2"

const LocalsSuccess = "success"

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Ok(c *fiber.Ctx, data interface{}) error {
	c.Locals(LocalsSuccess, true)
	return c.JSON(Envelope{Success: true, Data: data, Message: ""})
}

func OkMsg(c *fiber.Ctx, data interface{}, msg string) error {
	c.Locals(LocalsSuccess, true)
	return c.JSON(Envelope{Success: true, Data: data, Message: msg})
}

func Fail(c *fiber.Ctx, msg string) error {
	c.Locals(LocalsSuccess, false)
	return c.JSON(Envelope{Success: false, Data: nil, Message: msg})
}
