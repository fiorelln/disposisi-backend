package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
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
		Jabatans []uint `json:"jabatans"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(input.Jabatans) == 0 {
		c.JSON(400, gin.H{
			"error": "Jabatan wajib dipilih",
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
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(400, gin.H{
			"error": "Email sudah digunakan",
		})
		return
	}

	for index, jabatanID := range input.Jabatans {

		userJabatan := models.UserJabatan{
			UserID:    user.ID,
			JabatanID: jabatanID,
			IsPrimary: index == 0,
		}

		config.DB.Create(&userJabatan)
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
		Preload("Jabatans.Jabatan").
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

	var roles []string

	for _, jabatan := range user.Jabatans {
		roles = append(roles, jabatan.Jabatan.NamaJabatan)
	}

	token, err := helpers.GenerateToken(user.ID, roles)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal generate token",
		})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"jabatans": roles,
		},
	})
}

func ForgotPassword(c *gin.Context) {

	var input struct {
		Email string `json:"email"`
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

		c.JSON(404, gin.H{
			"error": "Email tidak ditemukan",
		})
		return
	}

	var total int64

	config.DB.
	Model(&models.OTP{}).
	Where("user_id = ? AND created_at > ?", user.ID, time.Now().Add(-25*time.Minute)).
	Count(&total)

	if total >= 5 {
		c.JSON(429, gin.H{
			"error": "Limit OTP tercapai",
		})
		return
	}

	rand.Seed(time.Now().UnixNano())

	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))

	otp := models.OTP{
		UserID:    user.ID,
		KodeOTP:   otpCode,
		ExpiresAt: time.Now().Add(2 * time.Minute),
		IsUsed:    false,
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal menyimpan OTP",
		})
		return
	}

	payload := map[string]interface{}{
		"from": os.Getenv("RESEND_EMAIL"),
		"to": []string{
			user.Email,
		},
		"subject": "Reset Password OTP",
		"html": "<h1>Kode OTP Kamu: " + otpCode + "</h1>",
	}

	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"POST",
		"https://api.resend.com/emails",
		bytes.NewBuffer(jsonPayload),
	)

	req.Header.Set("Authorization", "Bearer "+os.Getenv("RESEND_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal mengirim OTP",
		})
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		c.JSON(500, gin.H{
			"error": "Gagal mengirim OTP",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "OTP berhasil dikirim",
	})
}

func VerifyOTP(c *gin.Context) {

	var input struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
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

		c.JSON(404, gin.H{
			"error": "Email tidak ditemukan",
		})
		return
	}

	var otp models.OTP

	if err := config.DB.
		Where("user_id = ? AND kode_otp = ? AND is_used = ?", user.ID, input.OTP, false).
		Last(&otp).Error; err != nil {

		c.JSON(400, gin.H{
			"error": "OTP salah",
		})
		return
	}

	if time.Now().After(otp.ExpiresAt) {

		c.JSON(400, gin.H{
			"error": "OTP expired",
		})
		return
	}

	otp.IsUsed = true

	config.DB.Save(&otp)

	c.JSON(200, gin.H{
		"message": "OTP valid",
	})
}

func ResetPassword(c *gin.Context) {

	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !ValidatePassword(input.NewPassword) {
		c.JSON(400, gin.H{
			"error": "Password minimal 8 karakter, huruf besar, kecil, dan angka",
		})
		return
	}

	var user models.User

	if err := config.DB.
		Where("email = ?", input.Email).
		First(&user).Error; err != nil {

		c.JSON(404, gin.H{
			"error": "Email tidak ditemukan",
		})
		return
	}

	hashedPassword, err := helpers.HashPassword(input.NewPassword)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Gagal hash password",
		})
		return
	}

	user.Password = hashedPassword

	config.DB.Save(&user)

	c.JSON(200, gin.H{
		"message": "Password berhasil direset",
	})
}