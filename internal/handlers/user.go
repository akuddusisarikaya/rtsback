package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"github.com/dgrijalva/jwt-go"
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

var jwtKey = []byte("your_secret_key") // JWT için gizli anahtar

func Login(w http.ResponseWriter, r *http.Request) {
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

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Geçersiz şifre", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   user.ID.Hex(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token oluşturulamadı", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "E-posta belirtilmedi", http.StatusBadRequest)
		return
	}

	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		http.Error(w, "Kullanıcı bulunamadı veya veritabanı hatası", http.StatusNotFound)
		return
	}

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

func UpdateUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users []models.User
	err := json.NewDecoder(r.Body).Decode(&users)
	if err != nil {
		http.Error(w, "Veri çözümleme hatası: Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, user := range users {
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

		opts := options.Update().SetUpsert(true)
		_, err := userCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			log.Printf("Güncelleme hatası: %v", err)
			http.Error(w, "Veritabanı güncelleme hatası", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Kullanıcılar başarıyla güncellendi"})
}

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"email": user.Email}
	update := bson.M{"$set": user}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Database update error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

func SuperUserLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into creds struct
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	// Find the user by email
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = userCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Check if the password matches
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Check if the user is a SuperUser
	if !user.SuperUser {
		http.Error(w, "Access denied: User is not a SuperUser", http.StatusForbidden)
		return
	}

	// Create JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   user.ID.Hex(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("super_user_secret_key"))
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	// Return the token
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}