package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func getQueryInt(c *gin.Context, key string, defaultVal int) int {
	valStr := c.Query(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getOperatorInfo(c *gin.Context) (string, string) {
	opId := c.GetHeader("X-Operator-Id")
	if opId == "" {
		opId = "HR_Admin"
	}
	opName := c.GetHeader("X-Operator-Name")
	if opName == "" {
		opName = opId
	}
	return opId, opName
}
