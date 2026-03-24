package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var jwtKey = []byte("averysecretsecretkey")
var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr: "redis:6379",
})

// ========================
// JWT HELPER
// ========================
func getUserFromJWT(r *http.Request) (string, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return "", err
	}
	tokenStr := cookie.Value
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token")
	}
	return username, nil
}

// ========================
// COMMON LAYOUT
// ========================
func renderPage(w http.ResponseWriter, title string, body string) {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "dev"
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>%s</title>
		<style>
			body { font-family: Arial; text-align: center; margin-top: 50px; }
			a { display: inline-block; margin: 10px; padding: 10px 20px;
			    text-decoration: none; background-color: #333;
			    color: white; border-radius: 5px; }
			a:hover { background-color: #555; }
		</style>
	</head>
	<body>
		<h1>Backend Service (v%s)</h1>
		%s
	</body>
	</html>
	`, title, version, body)
}

// ========================
// BACKEND HANDLER
// ========================
func backendHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromJWT(r)
	if err != nil {
		// Redirect to public URL handled by reverse proxy
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	group, _ := rdb.Get(ctx, "group:"+user).Result()

	body := fmt.Sprintf(`
	<h2>Welcome %s</h2>
	<p>Your sticky group: %s</p>
	<a href="/logout">Logout</a>
	`, user, group)

	renderPage(w, "Backend Home", body)
}

// ========================
// MAIN
// ========================
func main() {
	http.HandleFunc("/backend", backendHandler)
	fmt.Println("Backend running on :3001")
	http.ListenAndServe(":3001", nil)
}
