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
func GetProviderAppointments(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    email := r.URL.Query().Get("email")
    date := r.URL.Query().Get("date")

    if email == "" || date == "" {
        http.Error(w, "Email and date are required", http.StatusBadRequest)
        return
    }

    // Parse the date from the request
    parsedDate, err := time.Parse("2006-01-02", date)
    if err != nil {
        http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    // Define the start and end of the day for the date
    startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
    endOfDay := startOfDay.Add(24 * time.Hour)

    // Create a filter to find appointments within the specified date range
    filter := bson.M{
        "provider_email": email,
        "date": bson.M{
            "$gte": startOfDay,
            "$lt":  endOfDay,
        },
    }

    // Fetch appointments from the database
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := appointmentCollection.Find(ctx, filter)
    if err != nil {
        http.Error(w, "Failed to fetch appointments", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    // Decode the fetched appointments into a slice
    var appointments []models.Appointment
    if err := cursor.All(ctx, &appointments); err != nil {
        http.Error(w, "Failed to decode appointments", http.StatusInternalServerError)
        return
    }

    // Respond with the appointments in JSON format
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(appointments)
}

func AddProviderApp(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Temporary struct for decoding JSON
    var appointmentData struct {
        ProviderEmail string `json:"providerEmail"`
        CompanyName   string `json:"companyName"`
        Date          string `json:"date"`       // Tarih string olarak alınıyor
        StartTime     string `json:"startTime"`  // Başlangıç saati string olarak alınıyor
        EndTime       string `json:"endTime"`    // Bitiş saati string olarak alınıyor
        Activate      bool   `json:"activate"`
    }

    // Decode the JSON body into appointmentData
    err := json.NewDecoder(r.Body).Decode(&appointmentData)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Parse the date and time fields
    parsedDate, err := time.Parse("2006-01-02", appointmentData.Date)
    if err != nil {
        http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
        return
    }

    parsedStartTime, err := time.Parse("15:04", appointmentData.StartTime)
    if err != nil {
        http.Error(w, "Invalid start time format, expected HH:MM", http.StatusBadRequest)
        return
    }

    parsedEndTime, err := time.Parse("15:04", appointmentData.EndTime)
    if err != nil {
        http.Error(w, "Invalid end time format, expected HH:MM", http.StatusBadRequest)
        return
    }

    // Create the appointment object
    appointment := models.Appointment{
        ID:            primitive.NewObjectID(),
        ProviderEmail: appointmentData.ProviderEmail,
        CompanyName:   appointmentData.CompanyName,
        Date:          parsedDate,
        StartTime:     parsedStartTime,
        EndTime:       parsedEndTime,
        Activate:      appointmentData.Activate,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Insert the appointment into the database
    _, err = appointmentCollection.InsertOne(ctx, appointment)
    if err != nil {
        http.Error(w, "Failed to add appointment to database", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Appointment added successfully"})
}

func UpdateAppointmentByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Appointment ID'yi URL parametresinden alın
	appointmentID := r.URL.Query().Get("id")
	if appointmentID == "" {
		http.Error(w, "Appointment ID is required", http.StatusBadRequest)
		return
	}

	// Appointment ID'yi ObjectID'ye çevirin
	objID, err := primitive.ObjectIDFromHex(appointmentID)
	if err != nil {
		http.Error(w, "Invalid Appointment ID", http.StatusBadRequest)
		return
	}

	// Gelen veriyi decode et
	var updateData struct {
  // Tarih string olarak alınıyor
        StartTime     string `json:"startTime"`  // Başlangıç saati string olarak alınıyor
        EndTime       string `json:"endTime"`    // Bitiş saati string olarak alınıyor
    }
	
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

    parsedStartTime, err := time.Parse("15:04", updateData.StartTime)
    if err != nil {
        http.Error(w, "Invalid start time format, expected HH:MM", http.StatusBadRequest)
        return
    }

    parsedEndTime, err := time.Parse("15:04", updateData.EndTime)
    if err != nil {
        http.Error(w, "Invalid end time format, expected HH:MM", http.StatusBadRequest)
        return
    }

    // Create the appointment object
    updatement := models.Appointment{
        StartTime:     parsedStartTime,
        EndTime:       parsedEndTime,
		UpdatedAt:     time.Now(),
    }

	// MongoDB güncelleme işlemi
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updatement}

	_, err = appointmentCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Failed to update appointment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Appointment updated successfully"})
}

func DeleteAppointmentByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Appointment ID'yi URL parametresinden alın
	appointmentID := r.URL.Query().Get("id")
	if appointmentID == "" {
		http.Error(w, "Appointment ID is required", http.StatusBadRequest)
		return
	}

	// Appointment ID'yi ObjectID'ye çevirin
	objID, err := primitive.ObjectIDFromHex(appointmentID)
	if err != nil {
		http.Error(w, "Invalid Appointment ID", http.StatusBadRequest)
		return
	}

	// MongoDB silme işlemi
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}

	_, err = appointmentCollection.DeleteOne(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to delete appointment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Appointment deleted successfully"})
}