package middlewares

import (
    "net/http"
    "strings"

    "github.com/dgrijalva/jwt-go"
)

func JwtVerify(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("your_secret_key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
func SuperUserJWT(next http.Handler) http.Handler{
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("super_user_secret_key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func AdminJWT(next http.Handler) http.Handler{
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("admin_secret_key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func ProviderJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        bearerToken := strings.Split(authHeader, " ")
        if len(bearerToken) != 2 {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("provider_secret_key"), nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
