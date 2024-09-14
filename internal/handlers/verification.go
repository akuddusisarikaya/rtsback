package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"rtsback/config"
	"rtsback/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/gomail.v2"
)

var verificationCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	verificationCollection = config.GetCollection(client, "verification")
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64()) // 6 haneli kod
}

func SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	userEmail := r.URL.Query().Get("email")
	userID := r.URL.Query().Get("userID")

	if userEmail == "" || userID == "" {
		http.Error(w, "Email and UserID are required", http.StatusBadRequest)
		return
	}

	code := generateCode()

	// Verification kaydı oluşturma
	verification := models.Verification{
		UserID:     userID,
		Email:      userEmail,
		EmailCode:  code,
		EmailVer:   false,
		CreateTime: time.Now(),
	}

	// Veritabanına kaydet
	_, err := verificationCollection.InsertOne(context.TODO(), verification)
	if err != nil {
		http.Error(w, "Failed to save verification code", http.StatusInternalServerError)
		return
	}

	// E-posta gönderimi
	err = sendVerificationEmail(userEmail, code)
	if err != nil {
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Verification code sent to your email!"))
	json.NewEncoder(w).Encode(map[string]string{"userID": verification.UserID})
}
func VerifyCode(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	code := r.URL.Query().Get("code")

	if userID == "" || code == "" {
		http.Error(w, "UserID and code are required", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı çek
	var verification models.Verification
	err := verificationCollection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&verification)
	if err != nil {
		http.Error(w, "Verification not found", http.StatusNotFound)
		return
	}

	// Kodu kontrol et
	if verification.EmailCode != code {
		http.Error(w, "Invalid verification code", http.StatusBadRequest)
		return
	}

	// Doğrulama işlemini güncelle
	update := bson.M{
		"$set": bson.M{
			"email_ver":      true,
			"email_ver_time": time.Now(),
		},
	}

	_, err = verificationCollection.UpdateOne(context.TODO(), bson.M{"user_id": userID}, update)
	if err != nil {
		http.Error(w, "Failed to update verification status", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Email verified successfully!"))
}

// E-posta gönderimi fonksiyonu
func sendVerificationEmail(email, code string) error {
	MailPassword := os.Getenv("MAIL_PASSWORD")
	m := gomail.NewMessage()
	m.SetHeader("From", "sahmedkuddusi@gmial.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Email Verification Code")
	m.SetBody("text/plain", fmt.Sprintf("Your verification code is: %s", code))

	// SMTP sunucu ayarları
	d := gomail.NewDialer("smtp.gmail.com", 587, "sahmedkuddusi@gmail.com", MailPassword)

	// E-posta gönderme işlemi
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		return err
	}
	return nil
}

func GetVerificationByUserIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den userID parametresini alıyoruz
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Veriyi çekmek için filtre oluşturma
	filter := bson.M{"user_id": userID}

	// Verification belgesini tutmak için boş bir değişken oluşturma
	var verification models.Verification

	// Belgeyi çekme
	err := verificationCollection.FindOne(context.TODO(), filter).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Verification data not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching verification data", http.StatusInternalServerError)
		return
	}

	// Verification verisini JSON formatında döner
	json.NewEncoder(w).Encode(verification)
}
