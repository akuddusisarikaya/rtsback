package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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

	providerID := provider.ID.Hex()

	// Başarılı yanıt ve token gönder
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "ID": providerID})
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

func GetAppointmentsByProviderEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query parametresinden email'i al
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter is required", http.StatusBadRequest)
		return
	}

	// MongoDB query için context oluştur
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Email'e göre filtreleme yap
	filter := bson.M{"provider_email": email}

	// Randevuları filtreye göre bul
	cursor, err := appointmentCollection.Find(ctx, filter, options.Find())
	if err != nil {
		http.Error(w, "Failed to fetch appointments", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Randevuları bir slice'a decode et
	var appointments []models.Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		http.Error(w, "Failed to decode appointments", http.StatusInternalServerError)
		return
	}

	// Randevuları JSON formatında yanıtla
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(appointments)
}

func GetProvidersByCompanyId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `companyID` parametresini al
	companyID := r.URL.Query().Get("companyID")
	if companyID == "" {
		http.Error(w, "Company ID is required", http.StatusBadRequest)
		return
	}

	// Sağlayıcıları veritabanından çek
	var providers []models.Provider
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// `companyID`'yi string olarak kullanarak sorgu yapıyoruz
	cursor, err := providerCollection.Find(ctx, bson.M{"company_id": companyID})
	if err != nil {
		http.Error(w, "Failed to fetch providers", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Cursor'u slice'a dekode et
	if err = cursor.All(ctx, &providers); err != nil {
		http.Error(w, "Failed to decode providers", http.StatusInternalServerError)
		return
	}

	// Sağlayıcıları JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(providers)
}

func AddServiceToProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// `providerID` parametresini al
	providerID := r.URL.Query().Get("providerID")
	if providerID == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	// ObjectID'yi doğrula
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		http.Error(w, "Invalid Provider ID", http.StatusBadRequest)
		return
	}

	type ServiceRequest struct {
		Services []string `json:"services"`
	}

	var serviceReq ServiceRequest

	// İstek gövdesini çözümle
	err = json.NewDecoder(r.Body).Decode(&serviceReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mevcut sağlayıcıyı veritabanından bul
	var provider models.Provider
	err = providerCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Yeni hizmetleri mevcut hizmetler ile birleştir
	existingServices := provider.Services
	for _, newService := range serviceReq.Services {
		if !contains(existingServices, newService) { // Eğer mevcut değilse ekle
			existingServices = append(existingServices, newService)
		}
	}

	// Güncellemeyi MongoDB'ye gönder
	update := bson.M{
		"$set": bson.M{
			"services":   existingServices,
			"updated_at": time.Now(),
		},
	}

	result, err := providerCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update provider services", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Service added successfully to the provider"})
}

func GetServicesOfProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `providerID` parametresini al
	providerID := r.URL.Query().Get("providerID")
	if providerID == "" {
		http.Error(w, "Provider ID is required", http.StatusBadRequest)
		return
	}

	// `providerID` string'ini `ObjectID`'ye çevir
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		http.Error(w, "Invalid Provider ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Veritabanından provider'ı çek
	var provider models.Provider
	err = providerCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Servisleri indeksleriyle birlikte hazırlamak için bir dizi oluştur
	type ServiceWithIndex struct {
		Index   int    `json:"index"`
		Service string `json:"service"`
	}

	var servicesWithIndex []ServiceWithIndex
	for i, service := range provider.Services {
		servicesWithIndex = append(servicesWithIndex, ServiceWithIndex{
			Index:   i,
			Service: service,
		})
	}

	// Servisleri JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(servicesWithIndex)
}

func RemoveServiceFromProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `providerID` ve `index` parametrelerini al
	providerID := r.URL.Query().Get("providerID")
	indexStr := r.URL.Query().Get("index")

	if providerID == "" || indexStr == "" {
		http.Error(w, "Provider ID and index are required", http.StatusBadRequest)
		return
	}

	// Provider ID'yi ObjectID'ye çevir
	objID, err := primitive.ObjectIDFromHex(providerID)
	if err != nil {
		http.Error(w, "Invalid Provider ID", http.StatusBadRequest)
		return
	}

	// Index'i integer olarak parse et
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Provider'ı veritabanından bul
	var provider struct {
		Services []string `bson:"services"`
	}

	err = providerCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&provider)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Eğer index geçersizse, hata döndür
	if index >= len(provider.Services) {
		http.Error(w, "Index out of range", http.StatusBadRequest)
		return
	}

	// Belirtilen indexteki öğeyi sil
	provider.Services = append(provider.Services[:index], provider.Services[index+1:]...)

	// Güncellenmiş services dizisini veritabanına kaydet
	update := bson.M{
		"$set": bson.M{"services": provider.Services},
	}

	_, err = providerCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update provider services", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Service removed successfully"})
}
