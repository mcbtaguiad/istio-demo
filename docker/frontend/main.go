package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
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
			form { margin: 20px; }
			input { display: block; margin: 10px auto; padding: 8px; }
			button { padding: 10px 20px; }
			a { display: inline-block; margin: 10px; padding: 10px 20px;
			    text-decoration: none; background-color: #333;
			    color: white; border-radius: 5px; }
			a:hover { background-color: #555; }
			.nav { margin-bottom: 20px; }
		</style>
	</head>
	<body>
		<h1>Frontend Service (v%s)</h1>
		<div class="nav">
			<a href="/">Home</a>
		</div>
		%s
	</body>
	</html>
	`, title, version, body)
}

// ========================
// HANDLERS
// ========================
func homeHandler(w http.ResponseWriter, r *http.Request) {
	body := `
	<h2>Welcome to the Landing Page</h2>
	<a href="/login">Login</a>
	<a href="/register">Register</a>
	`
	renderPage(w, "Home", body)
}

// ========================
// REGISTER
// ========================
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderPage(w, "Register", `
		<h2>Register</h2>
		<form method="POST" action="/register">
			<input name="username" placeholder="Username" required />
			<input name="password" type="password" placeholder="Password" required />
			<button type="submit">Register</button>
		</form>
		<a href="/login">Login</a>
		`)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	exists, _ := rdb.Exists(ctx, "user:"+username).Result()
	if exists == 1 {
		renderPage(w, "Error", "<p>User already exists</p><a href='/register'>Try again</a>")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	rdb.Set(ctx, "user:"+username, hash, 0)

	http.Redirect(w, r, "/login", http.StatusFound)
}

// ========================
// LOGIN
// ========================
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderPage(w, "Login", `
		<h2>Login</h2>
		<form method="POST" action="/login">
			<input name="username" placeholder="Username" required />
			<input name="password" type="password" placeholder="Password" required />
			<button type="submit">Login</button>
		</form>
		<a href="/register">Register</a>
		`)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	stored, err := rdb.Get(ctx, "user:"+username).Bytes()
	if err != nil {
		renderPage(w, "Error", "<p>Invalid credentials</p><a href='/login'>Try again</a>")
		return
	}

	err = bcrypt.CompareHashAndPassword(stored, []byte(password))
	if err != nil {
		renderPage(w, "Error", "<p>Invalid credentials</p><a href='/login'>Try again</a>")
		return
	}

	// JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwtKey)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
	})

	// Sticky group (optional frontend redirect)
	groupKey := "group:" + username
	group, err := rdb.Get(ctx, groupKey).Result()
	if err != nil {
		if time.Now().UnixNano()%2 == 0 {
			group = "alpha"
		} else {
			group = "beta"
		}
		rdb.Set(ctx, groupKey, group, 0)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "group",
		Value: group,
		Path:  "/",
	})

	http.Redirect(w, r, "/backend", http.StatusFound)
}

// ========================
// LOGOUT
// ========================
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "group",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// ========================
// MAIN
// ========================
func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Frontend running on :3000")
	http.ListenAndServe(":3000", nil)
}
