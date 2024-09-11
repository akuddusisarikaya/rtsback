package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var managerCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	managerCollection = config.GetCollection(client, "manager")
}
func AddManager(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var manager models.Manager
	err := json.NewDecoder(r.Body).Decode(&manager)
	if err != nil {
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(manager.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	manager.Password = string(hashedPassword)

	// Assign a new ObjectID and timestamps
	manager.ID = primitive.NewObjectID()
	manager.CreatedAt = time.Now()
	manager.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = managerCollection.InsertOne(ctx, manager)
	if err != nil {
		http.Error(w, "Failed to add manager to the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Manager added successfully"})
}
func ManagerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	var manager models.Manager
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Email ile manager'ı bul
	err = managerCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&manager)
	if err != nil {
		http.Error(w, "Manager not found", http.StatusUnauthorized)
		return
	}

	// Şifreyi doğrula
	err = bcrypt.CompareHashAndPassword([]byte(manager.Password), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// JWT token oluşturma
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   manager.ID.Hex(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("manager_secret_key"))
	if err != nil {
		http.Error(w, "Token oluşturulamadı", http.StatusInternalServerError)
		return
	}

	managerID := manager.ID.Hex()
	// Başarılı yanıt ve token gönder
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "ID": managerID})
}

func GetManagersByCompanyId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `companyID` parametresini al
	companyID := r.URL.Query().Get("companyID")
	if companyID == "" {
		http.Error(w, "Company ID is required", http.StatusBadRequest)
		return
	}

	// Sağlayıcıları veritabanından çek
	var managers []models.Manager
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// `companyID`'yi string olarak kullanarak sorgu yapıyoruz
	cursor, err := managerCollection.Find(ctx, bson.M{"company_id": companyID})
	if err != nil {
		http.Error(w, "Failed to fetch providers", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Cursor'u slice'a dekode et
	if err = cursor.All(ctx, &managers); err != nil {
		http.Error(w, "Failed to decode providers", http.StatusInternalServerError)
		return
	}

	// Sağlayıcıları JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(managers)
}