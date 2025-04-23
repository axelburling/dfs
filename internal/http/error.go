package http

import (
	"strings"

	"github.com/axelburling/dfs/internal/log"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type Error struct {
	log *log.Logger
	nodeId string
}
func NewError(log *log.Logger, id string) *Error {
	return &Error{
		log: log,
		nodeId: id,
	}
}

func (e *Error) New(c fiber.Ctx, msg string, code int, err error) error {
	var fields []zap.Field
	fields = append(fields, zap.String("node", e.nodeId))
	fields = append(fields, zap.String("ContentType", string(c.Request().Header.ContentType())))
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	e.log.Error(msg, fields...)
	return c.Status(code).JSON(fiber.Map{
		"error": capitalizeFirstLetter(msg),
	})
}

func capitalizeFirstLetter(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}