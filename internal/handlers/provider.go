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

var providerCollection *mongo.Collection

func init() {
    client := config.ConnectDB()
    providerCollection = config.GetCollection(client, "provider")
}

// AddProvider handles adding a new provider to the database
func AddProvider(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var provider models.Provider
    err := json.NewDecoder(r.Body).Decode(&provider)
    if err != nil {
        http.Error(w, "Invalid data format", http.StatusBadRequest)
        return
    }

    // Hash the password before storing
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(provider.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    provider.Password = string(hashedPassword)

    // Assign a new ObjectID and timestamps
    provider.ID = primitive.NewObjectID()
    provider.CreatedAt = time.Now()
    provider.UpdatedAt = time.Now()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = providerCollection.InsertOne(ctx, provider)
    if err != nil {
        http.Error(w, "Failed to add provider to the database", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Provider added successfully"})
}

// ProviderLogin handles provider login requests
func ProviderLogin(w http.ResponseWriter, r *http.Request) {
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

    var provider models.Provider
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Email ile provider'ı bul
    err = providerCollection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&provider)
    if err != nil {
        http.Error(w, "Provider not found", http.StatusUnauthorized)
        return
    }

    // Şifreyi doğrula
    err = bcrypt.CompareHashAndPassword([]byte(provider.Password), []byte(creds.Password))
    if err != nil {
        http.Error(w, "Invalid password", http.StatusUnauthorized)
        return
    }

    // JWT token oluşturma
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &jwt.StandardClaims{
        ExpiresAt: expirationTime.Unix(),
        Subject:   provider.ID.Hex(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("provider_secret_key"))
	if err != nil {
		http.Error(w, "Token oluşturulamadı", http.StatusInternalServerError)
		return
	}

    // Başarılı yanıt ve token gönder
    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func GetProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all providers
	cursor, err := providerCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch providers", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var providers []models.Provider
	for cursor.Next(ctx) {
		var provider models.Provider
		if err := cursor.Decode(&provider); err != nil {
			http.Error(w, "Failed to decode provider", http.StatusInternalServerError)
			return
		}
		providers = append(providers, provider)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	// Return providers as JSON
	if err := json.NewEncoder(w).Encode(providers); err != nil {
		http.Error(w, "Failed to encode providers to JSON", http.StatusInternalServerError)
		return
	}
}

func GetProviderByEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	var provider models.Provider
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := providerCollection.FindOne(ctx, bson.M{"email": email}).Decode(&provider)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Provider not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch provider", http.StatusInternalServerError)
		}
		return
	}

	// Successfully found the provider, send it in response
	json.NewEncoder(w).Encode(provider)
}

func GetCompanyNameByProviderEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query parametrelerinden e-posta değerini al
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is missing", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Sağlayıcıyı email üzerinden bul
	var provider models.Provider
	err := providerCollection.FindOne(ctx, bson.M{"email": email}).Decode(&provider)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Şirket bilgilerini JSON formatında yanıt olarak gönder
	json.NewEncoder(w).Encode(provider.CompanyName)
}