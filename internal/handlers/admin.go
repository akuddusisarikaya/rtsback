package handlers

import (
	"context"
	"encoding/json"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var adminCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	adminCollection = config.GetCollection(client, "admin")
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

func AddAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var admin models.Admin
	err := json.NewDecoder(r.Body).Decode(&admin)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	admin.ID = primitive.NewObjectID() // Yeni bir ObjectID oluşturur
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func AdminPanel(w http.ResponseWriter, r *http.Request) {

}

func GetPrices(w http.ResponseWriter, r *http.Request) {

}
