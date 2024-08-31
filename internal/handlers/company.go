package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"rtsback/config"
	"rtsback/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var companyCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	companyCollection = config.GetCollection(client, "company")
}
func GetCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var companies []models.Company
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := companyCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Veri çekme hatası", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &companies); err != nil {
		http.Error(w, "Veri çözümleme hatası", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(companies)
}

func AddCompany(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var company models.Company
	err := json.NewDecoder(r.Body).Decode(&company)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	company.ID = primitive.NewObjectID() // Yeni bir ObjectID oluşturur
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = companyCollection.InsertOne(ctx, company)
	if err != nil {
		http.Error(w, "Veritabanına eklenemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Şirket başarıyla eklendi"})
}

func UpdateCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var companies []models.Company
	err := json.NewDecoder(r.Body).Decode(&companies)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, company := range companies {
		filter := bson.M{"_id": company.ID}
		update := bson.M{"$set": company}

		_, err := companyCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			http.Error(w, "Veritabanı güncelleme hatası", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Şirketler başarıyla güncellendi"})
}
