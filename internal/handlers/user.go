package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	userCollection = config.GetCollection(client, "user")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Şifre hash'lemesi başarısız oldu", http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hashedPassword)

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Veritabanına eklenemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kullanıcıyı e-posta ile veritabanından bul
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"email": loginData.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı veya veritabanı hatası", http.StatusUnauthorized)
		return
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginData.Password))
	if err != nil {
		http.Error(w, "Şifre yanlış", http.StatusUnauthorized)
		return
	}

	// Giriş başarılı
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Kullanıcının e-posta adresini sorgu parametresinden al
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "E-posta belirtilmedi", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Kullanıcıyı veritabanında ara
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Println("Database Error: Kullanıcı bulunamadı veya veritabanı hatası:", err)
		http.Error(w, "Kullanıcı bulunamadı veya veritabanı hatası", http.StatusUnauthorized)
		return
	}

	// Kullanıcıyı JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users []models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := userCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Veri çekme hatası", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		http.Error(w, "Veri çözümleme hatası", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

// handlers/user.go

func UpdateUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Gelen veriyi çözümle
	var users []models.User
	err := json.NewDecoder(r.Body).Decode(&users)
	if err != nil {
		http.Error(w, "Veri çözümleme hatası: Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Her kullanıcı için güncelleme işlemi
	for _, user := range users {
		// Güncelleme için filtre ve update değerleri oluşturuluyor
		filter := bson.M{"_id": user.ID}
		update := bson.M{"$set": bson.M{
			"name":         user.Name,
			"email":        user.Email,
			"passwordHash": user.PasswordHash,
			"role":         user.Role,
			"phone":        user.Phone,
			"company_id":   user.CompanyID,
			"super_user":   user.SuperUser,
			"updatedAt":    time.Now(),
		}}

		// Güncelleme işlemini gerçekleştirme
		opts := options.Update().SetUpsert(true) // Eğer kayıt yoksa oluşturabilir
		_, err := userCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Güncelleme hatası: %v", err)
			http.Error(w, "Veritabanı güncelleme hatası", http.StatusInternalServerError)
			return
		}
	}

	// Başarılı güncelleme yanıtı
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Kullanıcılar başarıyla güncellendi"})
}
