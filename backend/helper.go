package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
	"os"
	"webrpl/table"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func isEmailValid(e string) bool {
    _, err := mail.ParseAddress(e)
    return err == nil
}

func checkOrMakeAdmin(backend *Backend, secret string) bool {
    reserved := "admin@wowadmin.com"
    var user table.User

    res := backend.db.Where("user_email = ?", reserved).First(&user)
    if res.Error == nil {
        if !CheckPassword(user.UserPassword, secret) {
            hashed, err := HashPassword(secret)
            if err != nil {
                return false
            }
            user.UserPassword = hashed
            if err := backend.db.Save(&user).Error; err != nil {
                return false
            }
        }
        return true
    }

    if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
        return false
    }

    hashed, err := HashPassword(secret)
    if err != nil {
        return false
    }

    user = table.User{
        UserEmail:    reserved,
        UserFullName: "admin",
        UserPassword: hashed,
        UserRole:     1,
    }

    if err := backend.db.Create(&user).Error; err != nil {
        return false
    }

    return true
}

func getCredentialFromEnv() string {
    password := os.Getenv("SECRET_KEY")
    if password == "" {
        password = "secret"
    }
    return password
}

func HashPassword(password string) (string, error) {
    // The cost parameter determines how computationally expensive the hash is to calculate
    // The default is 10, but you can increase it for better security (at the cost of performance)
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hashedBytes), nil
}

func CheckPassword(hashedPassword, plainPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
    return err == nil
}

func RandStringBytes(n int, rand_t *rand.Rand) string {
    const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func GetJWT(c *fiber.Ctx) (jwt.MapClaims, error) {
    user := c.Locals("user").(*jwt.Token)
    if user == nil {
        return nil, errors.New("JWT token not valid")
    }
    claims := user.Claims.(jwt.MapClaims)
    return claims, nil
}
