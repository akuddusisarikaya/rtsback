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
	protected := r.PathPrefix("/protected").Subrouter()
	protected.Use(middlewares.JwtVerify) // JWT doğrulama middleware'i ekle

	// Genel Rotlar
	r.HandleFunc("/user", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/adminlogin", handlers.LoginAdmin).Methods("POST")
	r.HandleFunc("/appointment", handlers.CreateAppointment).Methods("POST")
	r.HandleFunc("/prices", handlers.GetPrices).Methods("GET")

	// Korumalı Rotlar
	protected.HandleFunc("/userprofile", handlers.GetUserProfile).Methods("GET")
	protected.HandleFunc("/appointments", handlers.GetAppointments).Methods("GET")
	protected.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	protected.HandleFunc("/users/update", handlers.UpdateUsers).Methods("PUT") // Aynı endpoint üzerinde hem GET hem PUT olmamalı
	protected.HandleFunc("/admins", handlers.GetAdmins).Methods("GET")
	protected.HandleFunc("/companies", handlers.GetCompanies).Methods("GET")
	protected.HandleFunc("/companies", handlers.AddCompany).Methods("POST")
	protected.HandleFunc("/admins", handlers.AddAdmin).Methods("POST")
	protected.HandleFunc("/companies/update", handlers.UpdateCompanies).Methods("PUT") // Aynı endpoint üzerinde hem GET hem PUT olmamalı
	protected.HandleFunc("/admins/update", handlers.UpdateAdmins).Methods("PUT") // Aynı endpoint üzerinde hem GET hem PUT olmamalı
	protected.HandleFunc("/user/update", handlers.UpdateUserProfile).Methods("PUT") // Aynı endpoint üzerinde hem GET hem PUT olmamalı
	protected.HandleFunc("/admin", handlers.AdminPanel).Methods("GET")

	// CORS Ayarları
	corsRouter := middlewares.EnableCORS(r)

	// Sunucuyu başlat
	log.Println("Sunucu 8080 portunda çalışıyor...")
	if err := http.ListenAndServe(":8080", corsRouter); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
