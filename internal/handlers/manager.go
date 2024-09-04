package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "rtsback/config"
    "rtsback/internal/models"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "golang.org/x/crypto/bcrypt"
)

var managerCollection *mongo.Collection

func init() {
    client := config.ConnectDB()
    managerCollection = config.GetCollection(client, "manager")
}
func AddManager(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var manager models.Manager
    err := json.NewDecoder(r.Body).Decode(&manager)
    if err != nil {
        http.Error(w, "Invalid data format", http.StatusBadRequest)
        return
    }

    // Hash the password before storing
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(manager.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    manager.Password = string(hashedPassword)

    // Assign a new ObjectID and timestamps
    manager.ID = primitive.NewObjectID()
    manager.CreatedAt = time.Now()
    manager.UpdatedAt = time.Now()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err = managerCollection.InsertOne(ctx, manager)
    if err != nil {
        http.Error(w, "Failed to add managr to the database", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Manager added successfully"})
}
