package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("SUPER_SECRET_KEY")

func GenerateAccessTokenForFaculty(FacultyID string) (string, error) {
	claims := jwt.MapClaims{
		"faculty_id": FacultyID,
		"exp":        time.Now().Add(62 * 34 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAccessTokenForAdmin(AdminID string) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": AdminID,
		"exp":      time.Now().Add(62 * 34 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAccessTokenForWarden(WardenID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": WardenID,
		"exp":     time.Now().Add(62 * 34 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAccessTokenForMainAdmin(MainAdminID string) (string, error) {
	claims := jwt.MapClaims{
		"main_admin_id": MainAdminID,
		"exp":           time.Now().Add(62 * 34 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateAccessTokenForCollegeLogin(collegeEmail, collegeName, collegeId, superadminId string) (string, error) {
	claims := jwt.MapClaims{
		"super_admin_id": superadminId,
		"college_id":     collegeId,
		"college_name":   collegeName,
		"college_email":  collegeEmail,
		"exp":            time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// func GenerateRefreshToken(FacultyID string) (string, error) {
//     claims := jwt.MapClaims{
//         "faculty_id": FacultyID,
//         "exp":     time.Now().Add(31* 24 * time.Hour).Unix(),
//     }

//     token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//     return token.SignedString(jwtSecret)
// }
