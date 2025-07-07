package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Error struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statucCode int, message string) {
	logrus.Error(message)
	c.AbortWithStatusJSON(statucCode, Error{message})
}
