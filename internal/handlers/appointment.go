package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rtsback/config"
	"time"

	"rtsback/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var appointmentCollection *mongo.Collection

func init() {
	client := config.ConnectDB()
	appointmentCollection = config.GetCollection(client, "appointment")
}

// Tüm randevuları çeken handler
func GetAppointments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var appointments []models.Appointment
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := appointmentCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Veri çekme hatası", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &appointments); err != nil {
		http.Error(w, "Veri çözümleme hatası", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(appointments)
}

func CreateAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var appointment models.Appointment
	err := json.NewDecoder(r.Body).Decode(&appointment)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	appointment.ID = primitive.NewObjectID()
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = appointmentCollection.InsertOne(ctx, appointment)
	if err != nil {
		http.Error(w, "Veritabanına eklenemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Randevu başarıyla oluşturuldu"})
}