package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("SUPER_SECRET_KEY") 

func GenerateAccessToken(FacultyID string) (string, error) {
    claims := jwt.MapClaims{
        "faculty_id": FacultyID,
        "exp":     time.Now().Add(62*34 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func GenerateRefreshToken(FacultyID string) (string, error) {
    claims := jwt.MapClaims{
        "faculty_id": FacultyID,
        "exp":     time.Now().Add(31* 24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
