package main

import (
	"log"
	"net/http"
	"rtsback/config"
	"rtsback/internal/handlers"
	"rtsback/internal/middlewares"
	"rtsback/internal/models"

	"github.com/gorilla/mux"
)

func main() {
	client := config.ConnectDB()

	models.EnsureCollections(client)

	log.Println("Koleksiyonlar oluşturuldu ve sistem çalışıyor...")

	r := mux.NewRouter()

	r.HandleFunc("/user", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/userprofile", handlers.GetUserProfile).Methods("GET")
	r.HandleFunc("/users", handlers.UpdateUsers).Methods("PUT")

	r.HandleFunc("/admins", handlers.GetAdmins).Methods("GET")
	r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	r.HandleFunc("/companies", handlers.GetCompanies).Methods("GET")

	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/appointment", handlers.CreateAppointment).Methods("POST")
	r.HandleFunc("/admin", handlers.AdminPanel).Methods("GET")
	r.HandleFunc("/prices", handlers.GetPrices).Methods("GET")
	r.HandleFunc("/companies", handlers.AddCompany).Methods("POST")
	r.HandleFunc("/admins", handlers.AddAdmin).Methods("POST")
	r.HandleFunc("/companies", handlers.UpdateCompanies).Methods("PUT")
	r.HandleFunc("/admins", handlers.UpdateAdmins).Methods("PUT")


	corsRouter := middlewares.EnableCORS(r)

	// Sunucuyu başlat
	log.Println("Sunucu 8080 portunda çalışıyor...")
	if err := http.ListenAndServe(":8080", corsRouter); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
