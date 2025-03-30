package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	myjwt "github.com/hewo233/house-system-backend/utils/jwt"
	"log"
	"net/http"
)

func JWTAuth(audience string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			log.Println("No token")
			c.JSON(http.StatusBadRequest, gin.H{
				"errno":   40050,
				"message": "Bad Request, no token",
			})
			c.Abort()
			return
		}

		tokenString = tokenString[len("Bearer "):]

		token, err := jwt.ParseWithClaims(tokenString, &myjwt.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return myjwt.JWTKey, nil
		})
		if err != nil || !token.Valid {
			log.Println("Parse token error: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"errno":   50050,
				"message": "token parse error: " + err.Error(),
			})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*myjwt.Claims); ok {
			if claims.Audience != audience {
				log.Println("Audience error")
				c.JSON(http.StatusUnauthorized, gin.H{
					"errno": 40150,
					"msg":   "Unauthorized, audience error",
				})
				c.Abort()
				return
			}

			c.Set("phone", claims.StandardClaims.Id)
		}
	}
}
