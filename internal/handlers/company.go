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
	"go.mongodb.org/mongo-driver/mongo/options"
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


func GetCompanyByName(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Query parameterden şirket ismini al
    companyName := r.URL.Query().Get("name")
    if companyName == "" {
        http.Error(w, "Şirket adı gerekli", http.StatusBadRequest)
        return
    }

    var company models.Company
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // MongoDB'de şirketi isme göre ara
    err := companyCollection.FindOne(ctx, bson.M{"name": companyName}).Decode(&company)
    if err != nil {
        http.Error(w, "Şirket bulunamadı", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(company)
}

func UpdateCompanyByName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var updatedCompany models.Company
	err := json.NewDecoder(r.Body).Decode(&updatedCompany)
	if err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	// Şirket ismi boş ise hata döndür
	if updatedCompany.Name == "" {
		http.Error(w, "Şirket adı gerekli", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Şirket ismi ile güncelleme yap
	filter := bson.M{"name": updatedCompany.Name}
	update := bson.M{"$set": updatedCompany}

	opts := options.Update().SetUpsert(false) // Eğer şirket yoksa oluşturulmasın
	result, err := companyCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		http.Error(w, "Şirket güncellenemedi", http.StatusInternalServerError)
		return
	}

	// Hiçbir belge güncellenmediyse, hata mesajı döndür
	if result.MatchedCount == 0 {
		http.Error(w, "Şirket bulunamadı", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Şirket başarıyla güncellendi"})
}

func GetCompanyByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// URL'den `companyID` parametresini al
	companyID := r.URL.Query().Get("companyID")
	if companyID == "" {
		http.Error(w, "Company ID is required", http.StatusBadRequest)
		return
	}

	// `companyID` string'ini `ObjectID`'ye çevir
	objID, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		http.Error(w, "Invalid Company ID", http.StatusBadRequest)
		return
	}

	// Şirket bilgilerini veritabanından çek
	var company models.Company
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = companyCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&company)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Company not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch company information", http.StatusInternalServerError)
		return
	}

	// Şirket bilgilerini JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(company)
}

func GetAllCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var companies []models.Company
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Veritabanından tüm şirketleri çek
	cursor, err := companyCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Veri çekme hatası", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Cursor'u slice'a dekode et
	if err = cursor.All(ctx, &companies); err != nil {
		http.Error(w, "Veri çözümleme hatası", http.StatusInternalServerError)
		return
	}

	// Şirketleri JSON formatında döndür
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(companies)
}