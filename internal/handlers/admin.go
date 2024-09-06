package handlers

import (
	"context"
	"encoding/json"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"net/http"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var adminCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	adminCollection = config.GetCollection(client, "admin")
	userCollection = config.GetCollection(client, "user")
}

func LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	var admin models.Admin
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Admini e-posta ile bulma
	err = adminCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&admin)
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusUnauthorized)
		return
	}

	// Şifre doğrulama
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Geçersiz şifre", http.StatusUnauthorized)
		return
	}

	// Token oluşturma
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   admin.ID.Hex(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("admin_secret_key"))
	if err != nil {
		http.Error(w, "Token oluşturulamadı", http.StatusInternalServerError)
		return
	}
	adminID := (admin.ID).Hex()

	// Yanıt
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "ID": adminID})
}

// Admin verilerini çekme
func GetAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var admins []models.Admin
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := adminCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Veri çekme hatası", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &admins); err != nil {
		http.Error(w, "Veri çözümleme hatası", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(admins)
}

// Şifreyi hash'leyen fonksiyon
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func AddAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var admin models.Admin
	err := json.NewDecoder(r.Body).Decode(&admin)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	// Kullanıcıyı email üzerinden eşleştirin
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"email": admin.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusBadRequest)
		return
	}

	// Şifreyi hash'leyin
	hashedPassword, err := hashPassword(admin.Password)
	if err != nil {
		http.Error(w, "Şifre hash'lemesi başarısız oldu", http.StatusInternalServerError)
		return
	}
	admin.Password = hashedPassword

	// Adminin ek bilgilerini ayarlayın
	admin.ID = primitive.NewObjectID()
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()

	// Admin belgesini veritabanına ekleyin
	_, err = adminCollection.InsertOne(ctx, admin)
	if err != nil {
		http.Error(w, "Veritabanına eklenemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Admin başarıyla eklendi"})
}

func UpdateAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var admins []models.Admin
	err := json.NewDecoder(r.Body).Decode(&admins)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, admin := range admins {
		filter := bson.M{"_id": admin.ID}
		update := bson.M{"$set": admin}

		_, err := adminCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			http.Error(w, "Veritabanı güncelleme hatası", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Adminler başarıyla güncellendi"})
}


func GetAdminByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL parametrelerinden email'i al
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parametresi eksik", http.StatusBadRequest)
		return
	}

	// Admin bilgisini tutacak değişken
	var admin models.Admin

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Email ile admini bul
	err := adminCollection.FindOne(ctx, bson.M{"email": email}).Decode(&admin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Admin bulunamadı", http.StatusNotFound)
		} else {
			http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		}
		return
	}

	// Admin bilgilerini JSON formatında döner
	json.NewEncoder(w).Encode(admin)
}

func UpdateAdminByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL parametrelerinden email'i al
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parametresi eksik", http.StatusBadRequest)
		return
	}

	// Güncellenecek admin bilgisini al
	var updatedAdmin models.Admin
	err := json.NewDecoder(r.Body).Decode(&updatedAdmin)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Güncelleme verisi
	update := bson.M{
		"$set": bson.M{
			"name":      updatedAdmin.Name,
			"password":  updatedAdmin.Password, // Şifre güncelleniyorsa hash'lenmiş şekilde gelmeli
			"role":      updatedAdmin.Role,
			"updatedAt": time.Now(),
		},
	}

	// Email ile admini güncelle
	result, err := adminCollection.UpdateOne(ctx, bson.M{"email": email}, update)
	if err != nil {
		http.Error(w, "Veritabanı güncelleme hatası", http.StatusInternalServerError)
		return
	}

	// Eğer güncelleme yapılmadıysa
	if result.MatchedCount == 0 {
		http.Error(w, "Admin bulunamadı", http.StatusNotFound)
		return
	}

	// Başarılı yanıt
	json.NewEncoder(w).Encode(map[string]string{"message": "Admin başarıyla güncellendi"})
}