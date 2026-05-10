package controllers

import (
	"net/http"
	"regexp"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/helpers"
	"github.com/fiorelln/disposisi/models"

	"github.com/gin-gonic/gin"
)

var allowedRoles = map[string]bool{
	"super_admin":    true,
	"tu_admin":       true,
	"kepala_tu":      true,
	"kepsek":         true,
	"waka_kurikulum": true,
	"waka_kesiswaan": true,
	"waka_humas":     true,
	"waka_sarpras":   true,
	"bk":             true,
	"bkk":            true,
	"koordinator":    true,
	"prakerin":       true,
}

func ValidatePassword(password string) bool {

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return len(password) >= 8 &&
		hasUpper &&
		hasLower &&
		hasNumber
}

func Register(c *gin.Context) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !allowedRoles[input.Role] {
		c.JSON(400, gin.H{
			"error": "Role tidak valid",
		})
		return
	}

	if !ValidatePassword(input.Password) {
		c.JSON(400, gin.H{
			"error": "Password minimal 8 karakter, huruf besar, kecil, dan angka",
		})
		return
	}

	hashedPassword, err := helpers.HashPassword(input.Password)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal hash password",
		})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
		Role:     input.Role,
		Status:   "active",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(400, gin.H{
			"error": "Email sudah digunakan",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "User berhasil dibuat",
	})
}

func Login(c *gin.Context) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User

	if err := config.DB.
		Where("email = ?", input.Email).
		First(&user).Error; err != nil {

		c.JSON(401, gin.H{
			"error": "Email atau password salah",
		})
		return
	}

	if !helpers.CheckPassword(user.Password, input.Password) {
		c.JSON(401, gin.H{
			"error": "Email atau password salah",
		})
		return
	}

	token, err := helpers.GenerateToken(user.ID, user.Role)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal generate token",
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
