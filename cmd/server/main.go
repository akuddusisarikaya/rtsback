package main

import(
	"rtsback/config"
	"rtsback/internal/models"
	"log"
)

func main() {
    client := config.ConnectDB()

    models.EnsureCollections(client)


    log.Println("Koleksiyonlar oluşturuldu ve sistem çalışıyor...")
}