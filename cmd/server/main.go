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
	superuser := r.PathPrefix("/superuser").Subrouter()
	admin := r.PathPrefix("/admin").Subrouter()
	provider := r.PathPrefix("/provider").Subrouter()
	manager := r.PathPrefix("/manager").Subrouter()
	protected.Use(middlewares.JwtVerify) // JWT doğrulama middleware'i ekle
	admin.Use(middlewares.AdminJWT)
	superuser.Use(middlewares.SuperUserJWT)
	provider.Use(middlewares.ProviderJWT)
	manager.Use(middlewares.ManagerJWT)

	// Genel Rotlar
	r.HandleFunc("/user", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/provider/login", handlers.ProviderLogin).Methods("POST")
	r.HandleFunc("/adminlogin", handlers.LoginAdmin).Methods("POST")
	r.HandleFunc("/superuserlogin", handlers.SuperUserLogin).Methods("POST")
	r.HandleFunc("/managerlogin", handlers.ManagerLogin).Methods("POST")
	r.HandleFunc("/getproviderbycompany", handlers.GetProvidersByCompanyId).Methods("GET")
	r.HandleFunc("/getprovidersapp", handlers.GetInactiveAppointmentsOfProvider).Methods("GET")
	r.HandleFunc("/getallcompanies", handlers.GetAllCompanies).Methods("GET")
	r.HandleFunc("/makeappointment", handlers.UpdateAppointmentFieldsByID).Methods("PUT")
	r.HandleFunc("/getuserbyemail", handlers.GetUserByEmail).Methods("GET")
	r.HandleFunc("/createusernopassword", handlers.CreateUserWithoutPassword).Methods("POST")

	// Korumalı Rotlar
	admin.HandleFunc("/provider/add", handlers.AddProvider).Methods("POST")
	provider.HandleFunc("/getbyemail", handlers.GetProviderByEmail).Methods("GET")
	provider.HandleFunc("/providers", handlers.GetProviders).Methods("GET")
	admin.HandleFunc("/manager/add", handlers.AddManager).Methods("POST")
	provider.HandleFunc("/addappauto", handlers.AutoCreateAppointment).Methods("POST")
	protected.HandleFunc("/userprofile", handlers.GetUserProfile).Methods("GET")
	protected.HandleFunc("/appointments", handlers.GetAppointments).Methods("GET")
	superuser.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	superuser.HandleFunc("/users/update", handlers.UpdateUsers).Methods("PUT")
	superuser.HandleFunc("/admins", handlers.GetAdmins).Methods("GET")
	superuser.HandleFunc("/companies", handlers.GetCompanies).Methods("GET")
	superuser.HandleFunc("/companies", handlers.AddCompany).Methods("POST")
	superuser.HandleFunc("/adminadd", handlers.AddAdmin).Methods("POST")
	superuser.HandleFunc("/company/update", handlers.UpdateCompanyByName).Methods("PUT")
	superuser.HandleFunc("/admins/update", handlers.UpdateAdminByEmail).Methods("PUT")
	superuser.HandleFunc("/adminsget", handlers.GetAdminByEmail).Methods("GET")
	admin.HandleFunc("/adminsget", handlers.GetAdminByEmail).Methods("GET")
	superuser.HandleFunc("/companyget", handlers.GetCompanyByName).Methods("GET")
	admin.HandleFunc("/user/update", handlers.UpdateUserProfile).Methods("PUT")
	provider.HandleFunc("/getappointments", handlers.GetProviderAppointments).Methods("GET")
	provider.HandleFunc("/getcompanyforprovider", handlers.GetCompanyNameByProviderEmail).Methods("GET")
	provider.HandleFunc("/addproviderapp", handlers.AddProviderApp).Methods("POST")
	provider.HandleFunc("/updateapp", handlers.UpdateAppointmentByID).Methods("PUT")
	provider.HandleFunc("/deleteapp", handlers.DeleteAppointmentByID).Methods("DELETE")
	provider.HandleFunc("/getallproviderapp", handlers.GetAppointmentsByProviderEmail).Methods("GET")
	admin.HandleFunc("/getallproviderapp", handlers.GetAppointmentsByProviderEmail).Methods("GET")
	admin.HandleFunc("/getcompanybyid", handlers.GetCompanyByID).Methods("GET")
	provider.HandleFunc("/addservices", handlers.AddServiceToProvider).Methods("PUT")
	provider.HandleFunc("/getservicesforprovider", handlers.GetServicesOfProvider).Methods("GET")
	provider.HandleFunc("/deleteservice", handlers.RemoveServiceFromProvider).Methods("DELETE")

	// CORS Ayarları
	corsRouter := middlewares.EnableCORS(r)

	// Sunucuyu başlat
	log.Println("Sunucu 8080 portunda çalışıyor...")
	if err := http.ListenAndServe(":8080", corsRouter); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
