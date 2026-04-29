package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("averysecretsecretkey")
var ctx = context.Background()
var rdb *redis.Client

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getVersion() string {
	v := os.Getenv("VERSION")
	if v == "" {
		return "dev"
	}
	return v
}

func getRedisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return "redis:6379"
	}
	return addr
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: getRedisAddr(),
	})
}

func generateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iss":      "demo-app",
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtKey)
}

func randomGroup() string {
	groups := []string{"alpha", "beta"}
	return groups[rand.Intn(len(groups))]
}

// Middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r.Header.Set("user", username)
		next(w, r)
	}
}

// Handlers

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	exists, _ := rdb.Exists(ctx, "user:"+creds.Username).Result()
	if exists == 1 {
		http.Error(w, "User exists", http.StatusConflict)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(creds.Password), 10)
	rdb.Set(ctx, "user:"+creds.Username, hash, 0)

	group := randomGroup()
	rdb.Set(ctx, "group:"+creds.Username, group, 0)

	json.NewEncoder(w).Encode(map[string]string{
		"group": group,
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	stored, err := rdb.Get(ctx, "user:"+creds.Username).Bytes()
	if err != nil || bcrypt.CompareHashAndPassword(stored, []byte(creds.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := generateJWT(creds.Username)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("user")
	group, _ := rdb.Get(ctx, "group:"+user).Result()

	json.NewEncoder(w).Encode(map[string]string{
		"username": user,
		"group":    group,
	})
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	keys, _ := rdb.Keys(ctx, "user:*").Result()

	type User struct {
		Username string `json:"username"`
		Group    string `json:"group"`
	}

	var users []User

	for _, k := range keys {
		username := k[len("user:"):]
		group, _ := rdb.Get(ctx, "group:"+username).Result()

		users = append(users, User{
			Username: username,
			Group:    group,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(users),
		"users": users,
	})
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/users/")
	deleted, _ := rdb.Del(ctx, "user:"+username, "group:"+username).Result()

	if deleted == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "user deleted",
	})
}

func updatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/api/users/")

	var body struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	exists, _ := rdb.Exists(ctx, "user:"+username).Result()
	if exists == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	rdb.Set(ctx, "user:"+username, hash, 0)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "password updated",
	})
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"version": getVersion(),
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := rdb.Ping(ctx).Err()

	status := "ok"
	if err != nil {
		status = "degraded"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"version":   getVersion(),
		"timestamp": time.Now().UTC(),
	})
}

func versionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Version", getVersion())
		next.ServeHTTP(w, r)
	})
}

func main() {
	rdb = initRedis()

	allowedHostsEnv := os.Getenv("ALLOWED_HOSTS")
	originMap := make(map[string]bool)

	// default
	originMap["http://localhost:3001"] = true

	// env
	if allowedHostsEnv != "" {
		for _, host := range strings.Split(allowedHostsEnv, ",") {
			host = strings.TrimSpace(host)
			if host != "" {
				originMap[host] = true
			}
		}
	}

	origins := make([]string, 0, len(originMap))
	for origin := range originMap {
		origins = append(origins, origin)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/register", registerHandler)
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/profile", authMiddleware(profileHandler))
	mux.HandleFunc("/api/version", versionHandler)
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/users", authMiddleware(listUsersHandler))
	mux.HandleFunc("/api/users/", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			deleteUserHandler(w, r)
		case http.MethodPut:
			updatePasswordHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
	})

	http.ListenAndServe(":3000", versionMiddleware(c.Handler(mux)))
}
