package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleErrors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		err := ctx.Errors.ByType(gin.ErrorTypePublic).Last()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
		}
	}
}
