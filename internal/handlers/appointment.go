package handlers

import (
	"encoding/json"
	"net/http"
)


func CreateAppointment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var appointmentData struct {
		UserID    string `json:"userId"`
		ServiceID string `json:"serviceId"`
		DateTime  string `json:"dateTime"`
	}

	err := json.NewDecoder(r.Body).Decode(&appointmentData)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Appointment created successfully"})
}
