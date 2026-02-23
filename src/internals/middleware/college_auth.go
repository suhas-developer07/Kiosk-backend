package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func CollegeAuth(logger *zap.SugaredLogger) echo.MiddlewareFunc {
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
			collegeID, ok := claims["college_id"].(string)
			if !ok {
				return unauthorized(c, "Invalid college_id in token")
			}

			superAdminID, ok := claims["super_admin_id"].(string)
			if !ok {
				return unauthorized(c, "Invalid super_admin_id in token")
			}

			collegeName, ok := claims["college_name"].(string)
			if !ok {
				return unauthorized(c, "Invalid college_name in token")
			}

			collegeEmail, ok := claims["college_email"].(string)
			if !ok {
				return unauthorized(c, "Invalid college_email in token")
			}

			c.Set("college_id", collegeID)
			c.Set("super_admin_id", superAdminID)
			c.Set("college_name", collegeName)
			c.Set("college_email", collegeEmail)

			return next(c)
		}
	}
}

func unauthorized(c echo.Context, msg string) error {
    return c.JSON(http.StatusUnauthorized, map[string]string{
        "status": "error",
        "error":  msg,
    })
}
