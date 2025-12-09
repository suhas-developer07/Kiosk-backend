package utils

import "golang.org/x/crypto/bcrypt"

func CheckPassword(raw, hashed string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw)) == nil
}
