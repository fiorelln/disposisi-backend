package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fiorelln/disposisi/config"
	"github.com/fiorelln/disposisi/helpers"
	"github.com/fiorelln/disposisi/models"

	"github.com/gin-gonic/gin"
)

func ValidatePassword(password string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	return len(password) >= 8 && hasUpper && hasLower && hasNumber
}

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(email)
}

func Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Jabatans []uint `json:"jabatans"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !ValidateEmail(input.Email) {
		c.JSON(400, gin.H{"error": "email invalid"})
		return
	}

	if !ValidatePassword(input.Password) {
		c.JSON(400, gin.H{"error": "password lemah"})
		return
	}

	hash, _ := helpers.HashPassword(input.Password)

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hash,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": "email sudah digunakan"})
		return
	}

	for i, j := range input.Jabatans {
		config.DB.Create(&models.UserJabatan{
			UserID:    user.ID,
			JabatanID: j,
			IsPrimary: i == 0,
		})
	}

	c.JSON(200, gin.H{"message": "user created"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(401, gin.H{"error": "login gagal"})
		return
	}

	if !helpers.CheckPassword(user.Password, input.Password) {
		c.JSON(401, gin.H{"error": "login gagal"})
		return
	}

	var userJabatans []models.UserJabatan
	config.DB.Where("id_user = ?", user.ID).Find(&userJabatans)

	roles := []string{}
	for _, uj := range userJabatans {
		var jabatan models.Jabatan
		if err := config.DB.First(&jabatan, uj.JabatanID).Error; err == nil {
			roles = append(roles, jabatan.NamaJabatan)
		}
	}

	token, _ := helpers.GenerateToken(user.ID, roles)

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "user tidak ditemukan"})
		return
	}

	var count int64
	config.DB.Model(&models.OTP{}).
		Where("id_user = ? AND created_at > ?", user.ID, time.Now().Add(-25*time.Minute)).
		Count(&count)

	if count >= 5 {
		c.JSON(429, gin.H{"error": "limit OTP 5x / 25 menit"})
		return
	}

	var last models.OTP
	err := config.DB.Where("id_user = ?", user.ID).
		Order("created_at desc").
		First(&last).Error

	if err == nil {
		if time.Since(last.CreatedAt) < 60*time.Second {
			c.JSON(429, gin.H{"error": "tunggu 60 detik"})
			return
		}
	}

	config.DB.Model(&models.OTP{}).
		Where("id_user = ? AND is_used = false", user.ID).
		Update("is_used", true)

	code := generateOTP()

	otp := models.OTP{UserID: user.ID,
		KodeOTP:   code,
		ExpiresAt: time.Now().Add(2 * time.Minute),
		IsUsed:    false,
	}

	config.DB.Create(&otp)

	payload := map[string]interface{}{
		"from":    os.Getenv("RESEND_EMAIL"),
		"to":      []string{user.Email},
		"subject": "OTP Reset Password",
		"html":    "<h2>Kode OTP: " + code + "</h2>",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"POST",
		"https://api.resend.com/emails",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+os.Getenv("RESEND_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode >= 300 {
		c.JSON(500, gin.H{"error": "gagal kirim otp"})
		return
	}

	c.JSON(200, gin.H{"message": "otp terkirim"})
}

func VerifyOTP(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "user tidak ditemukan"})
		return
	}

	var otp models.OTP
	if err := config.DB.
		Where("id_user = ? AND kode_otp = ? AND is_used = false", user.ID, input.OTP).
		Order("created_at desc").
		First(&otp).Error; err != nil {

		c.JSON(400, gin.H{"error": "otp salah"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(400, gin.H{"error": "otp expired"})
		return
	}

	otp.IsUsed = true
	config.DB.Save(&otp)

	c.JSON(200, gin.H{"message": "otp valid"})
}

func ResetPassword(c *gin.Context) {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if !ValidatePassword(input.NewPassword) {
		c.JSON(400, gin.H{"error": "password lemah"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "user tidak ditemukan"})
		return
	}

	var otp models.OTP
	if err := config.DB.
		Where("id_user = ? AND is_used = true", user.ID).
		Order("created_at desc").
		First(&otp).Error; err != nil {

		c.JSON(400, gin.H{"error": "otp belum diverifikasi"})
		return
	}

	if time.Since(otp.CreatedAt) > 5*time.Minute {
		c.JSON(400, gin.H{"error": "session expired"})
		return
	}

	hash, _ := helpers.HashPassword(input.NewPassword)

	user.Password = hash
	config.DB.Save(&user)

	c.JSON(200, gin.H{"message": "password berhasil direset"})
}