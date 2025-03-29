package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"manage-products/utils"
	"net/http"
	"strings"
)

func AuthenticateMiddleware(c *gin.Context) {
	tokenHeader := c.GetHeader("Authorization")
	if tokenHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "missing token",
		})
		c.Abort()
		return
	}

	splitToken := strings.Split(tokenHeader, " ")
	if len(splitToken) != 2 || !strings.EqualFold(splitToken[0], "Bearer") {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid bearer token",
		})
		c.Abort()
		return
	}

	token, err := utils.VerifyToken(splitToken[1])
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "invalid token",
		})
		c.Abort()
		return
	}

	fmt.Printf("Token verified successfully. Claims: %+v\\n", token.Claims)
	c.Next()
}
