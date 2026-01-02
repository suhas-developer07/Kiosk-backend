package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var jwtSecret = []byte("SUPER_SECRET_KEY")

func AuthMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"status": "error",
					"error":  "Missing Authorization header",
				})
			}

			// Expect "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Warn("Invalid Authorization header format")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"status": "error",
					"error":  "Invalid Authorization format",
				})
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.Error("Unexpected signing method")
					return nil, echo.ErrUnauthorized
				}
				return jwtSecret, nil
			})

			if err != nil {
				logger.Warnf("Invalid token: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"status": "error",
					"error":  "Invalid or expired token",
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				logger.Warn("Invalid token claims")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"status": "error",
					"error":  "Invalid token",
				})
			}

			FacultyID, ok := claims["faculty_id"].(string)
			if !ok {
				logger.Error("Missing user_id in token")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"status": "error",
					"error":  "Invalid token payload",
				})
			}

			c.Set("faculty_id", FacultyID)

			return next(c)
		}
	}
}
