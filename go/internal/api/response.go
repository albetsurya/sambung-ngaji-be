package api

import "github.com/gofiber/fiber/v2"

// Envelope response — KONTRAK INI HARUS DIPERTAHANKAN.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func Ok(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Envelope{Success: true, Data: data, Message: ""})
}

func OkMsg(c *fiber.Ctx, data interface{}, msg string) error {
	return c.JSON(Envelope{Success: true, Data: data, Message: msg})
}

func Fail(c *fiber.Ctx, msg string) error {
	return c.JSON(Envelope{Success: false, Data: nil, Message: msg})
}
