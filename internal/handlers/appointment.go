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
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
func AutoCreateAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var autoAdd models.AutoAddRequest
	err := json.NewDecoder(r.Body).Decode(&autoAdd)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Geçerli tarihi al ve bir ay sonrasını hesapla
	currentDate := time.Now()
	endDate := currentDate.AddDate(0, 1, 0) // Bir ay sonrası

	// Belirtilen günlerde ve saatlerde randevu oluşturma işlemi
	for currentDate.Before(endDate) {
		currentWeekday := currentDate.Weekday().String()

		// Eğer geçerli gün belirtilen günler arasında varsa, randevu oluştur
		if contains(autoAdd.Weekdays, currentWeekday) {
			startTime, _ := time.Parse("15:04", autoAdd.ShiftStart)
			endTime, _ := time.Parse("15:04", autoAdd.ShiftEnd)

			start := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
				startTime.Hour(), startTime.Minute(), 0, 0, currentDate.Location())

			end := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
				endTime.Hour(), endTime.Minute(), 0, 0, currentDate.Location())

			for start.Before(end) {
				appointment := models.Appointment{
					ID:            primitive.NewObjectID(),
					ProviderEmail: autoAdd.ProviderEmail,
					CompanyName:   autoAdd.CompanyName,
					Date:          start,
					StartTime:     start,
					EndTime:       start.Add(time.Duration(autoAdd.Period) * time.Minute),
					Activate:      false,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}

				// Randevuyu MongoDB'ye ekle
				_, err := appointmentCollection.InsertOne(ctx, appointment)
				if err != nil {
					http.Error(w, "Veritabanına randevu eklenemedi", http.StatusInternalServerError)
					return
				}

				// Başlangıç saatini periyot kadar artır
				start = start.Add(time.Duration(autoAdd.Period) * time.Minute)
			}
		}

		// Tarihi bir gün ileri al
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Otomatik randevular başarıyla oluşturuldu"})
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
		Date          string `json:"date"`      // Tarih string olarak alınıyor
		StartTime     string `json:"startTime"` // Başlangıç saati string olarak alınıyor
		EndTime       string `json:"endTime"`   // Bitiş saati string olarak alınıyor
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
		StartTime string `json:"startTime"` // Başlangıç saati string olarak alınıyor
		EndTime   string `json:"endTime"`   // Bitiş saati string olarak alınıyor
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
		StartTime: parsedStartTime,
		EndTime:   parsedEndTime,
		UpdatedAt: time.Now(),
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

func GetInactiveAppointmentsOfProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `providerEmail` ve `date` parametrelerini al
	providerEmail := r.URL.Query().Get("providerEmail")
	date := r.URL.Query().Get("date")

	if providerEmail == "" || date == "" {
		http.Error(w, "Provider Email and date are required", http.StatusBadRequest)
		return
	}

	// Tarihi parse et
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Tarihin başlangıcı ve bitişi
	startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Veritabanı sorgusu: providerEmail'e, tarihe ve activate=false olma durumuna göre
	filter := bson.M{
		"provider_email": providerEmail,
		"date": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
		"activate": false,
	}

	// Randevuları çek
	var appointments []models.Appointment
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := appointmentCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to fetch appointments", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Cursor'u slice'a dekode et
	if err = cursor.All(ctx, &appointments); err != nil {
		http.Error(w, "Failed to decode appointments", http.StatusInternalServerError)
		return
	}

	// Randevuları JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(appointments)
}

func UpdateAppointmentFieldsByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `appointmentID` parametresini al
	appointmentID := r.URL.Query().Get("appointmentID")
	if appointmentID == "" {
		http.Error(w, "Appointment ID is required", http.StatusBadRequest)
		return
	}

	// `appointmentID` string'ini `ObjectID`'ye çevir
	objID, err := primitive.ObjectIDFromHex(appointmentID)
	if err != nil {
		http.Error(w, "Invalid Appointment ID", http.StatusBadRequest)
		return
	}

	// Güncellenecek alanları almak için geçici bir struct tanımlayın
	var updateData struct {
		CustomerName  string   `json:"customer_name,omitempty"`
		CustomerEmail string   `json:"customer_email,omitempty"`
		Services      []string `json:"services,omitempty"`
		Activate      *bool    `json:"activate"`
	}

	// Gelen JSON verisini decode et
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Güncelleme için `UpdatedAt` alanını ayarlayın
	update := bson.M{
		"$set": bson.M{
			"customer_name":  updateData.CustomerName,
			"customer_email": updateData.CustomerEmail,
			"services":       updateData.Services,
			"activate":       updateData.Activate,
			"updated_at":     time.Now(),
		},
	}

	// MongoDB güncelleme işlemi
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := appointmentCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, "Failed to update appointment", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Appointment not found", http.StatusNotFound)
		return
	}

	// Başarı durumunda mesaj döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Appointment fields updated successfully"})
}
