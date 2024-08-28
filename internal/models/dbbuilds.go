package models

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
)

var collections = []string{"admin", "appointment", "auth", "company", "manager", "provider", "user","availebleappointment"}

func EnsureCollections(db *mongo.Client) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    database := db.Database("rtsdatabase")

    for _, colName := range collections {
        if err := CreateCollectionIfNotExists(database, colName, ctx); err != nil {
            log.Fatalf("Koleksiyon oluşturulamadı: %s, hata: %v", colName, err)
        }
    }

    log.Println("Tüm koleksiyonlar kontrol edildi ve varsa oluşturuldu.")
}

func CreateCollectionIfNotExists(database *mongo.Database, collectionName string, ctx context.Context) error {
    collections, err := database.ListCollectionNames(ctx, map[string]interface{}{"name": collectionName})
    if err != nil {
        return err
    }

    if len(collections) == 0 {
        log.Printf("Koleksiyon yok, oluşturuluyor: %s", collectionName)
        return database.CreateCollection(ctx, collectionName)
    }

    log.Printf("Koleksiyon zaten var: %s", collectionName)
    return nil
}
