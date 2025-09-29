// --- REST API GOLANG: Konfigurasi External & Security ---
// Sesuai spesifikasi workshop hari ke-3

package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// Idol merepresentasikan satu idol K-Pop
type Idol struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Group    string `json:"group"`
	Position string `json:"position"`
}

// User dari config.yaml
type User struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"-"`
}

type Config struct {
	Users []User `yaml:"users"`
}

// TokenInfo untuk menyimpan token aktif di memory
type TokenInfo struct {
	Username  string
	ExpiresAt time.Time
}

var (
	// In-memory storage
	idols = []Idol{
		{ID: "1", Name: "Jisung", Group: "NCT", Position: "Main Dancer"},
		{ID: "2", Name: "Karina", Group: "AESPA", Position: "Leader"},
		{ID: "3", Name: "Ahyeon", Group: "BABYMONSTER", Position: "Main Vocal"},
	}
	nextID = 4

	config     Config
	usersMap   = map[string]string{} // username:password
	tokenStore = struct {
		sync.RWMutex
		m map[string]TokenInfo
	}{m: make(map[string]TokenInfo)}
	tokenTTL  = time.Hour
	envPort   = "8080"
	basicUser = ""
	basicPass = ""
)

// --- UTILITAS KONFIGURASI ---
func loadConfig() error {
	// Load .env
	_ = godotenv.Load()
	if port := os.Getenv("PORT"); port != "" {
		envPort = port
	}
	basicUser = os.Getenv("BASIC_USER")
	basicPass = os.Getenv("BASIC_PASS")
	// Load config.yaml
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	for _, u := range config.Users {
		usersMap[u.Username] = u.Password
	}
	return nil
}

// --- TOKEN ---
func generateToken(username string) (string, TokenInfo) {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	info := TokenInfo{Username: username, ExpiresAt: time.Now().Add(tokenTTL)}
	tokenStore.Lock()
	tokenStore.m[token] = info
	tokenStore.Unlock()
	return token, info
}

func validateToken(token string) (TokenInfo, bool) {
	tokenStore.RLock()
	info, ok := tokenStore.m[token]
	tokenStore.RUnlock()
	if !ok || time.Now().After(info.ExpiresAt) {
		return TokenInfo{}, false
	}
	return info, true
}

func revokeToken(token string) {
	tokenStore.Lock()
	delete(tokenStore.m, token)
	tokenStore.Unlock()
}

// --- MIDDLEWARE ---
// CORS middleware untuk semua origin, method, dan header
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Bearer-Token")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Auth middleware: validasi token kecuali /api/login dan static
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/login") ||
			strings.HasPrefix(r.URL.Path, "/login.html") ||
			strings.HasPrefix(r.URL.Path, "/index.html") ||
			strings.HasPrefix(r.URL.Path, "/static/") ||
			r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		_, ok := validateToken(token)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		// Simpan username di context jika perlu
		next.ServeHTTP(w, r)
	})
}

// --- RESPONSE HELPER ---
func addCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Bearer-Token")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	addCORS(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// --- ENDPOINTS ---

// handleOptions menangani preflight/OPTIONS agar bisa diakses dari frontend
func handleOptions(w http.ResponseWriter, r *http.Request) {
	addCORS(w)
	w.WriteHeader(http.StatusNoContent)
}

// handleIdols mengelola endpoint GET dan POST pada /api/idols
func handleIdols(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET /api/idols -> semua idol
		writeJSON(w, http.StatusOK, idols)
	case http.MethodPost:
		// POST /api/idols -> tambah idol baru
		var in struct {
			Name     string `json:"name"`
			Group    string `json:"group"`
			Position string `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		id := strconv.Itoa(nextID)
		nextID++
		newIdol := Idol{ID: id, Name: in.Name, Group: in.Group, Position: in.Position}
		idols = append(idols, newIdol)
		writeJSON(w, http.StatusCreated, newIdol)
	case http.MethodOptions:
		handleOptions(w, r)
	default:
		handleOptions(w, r)
	}
}

// handleIdolByID mengelola endpoint PUT dan DELETE pada /api/idols/{id}
func handleIdolByID(w http.ResponseWriter, r *http.Request) {
	// Ekstrak ID dari path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/idols/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "missing id in path")
		return
	}
	id := parts[0]
	switch r.Method {
	case http.MethodPut:
		// PUT /api/idols/{id} -> update idol
		var in struct {
			Name     string `json:"name"`
			Group    string `json:"group"`
			Position string `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		for i := range idols {
			if idols[i].ID == id {
				idols[i].Name = in.Name
				idols[i].Group = in.Group
				idols[i].Position = in.Position
				writeJSON(w, http.StatusOK, idols[i])
				return
			}
		}
		writeError(w, http.StatusNotFound, "idol not found")
	case http.MethodDelete:
		// DELETE /api/idols/{id} -> hapus idol
		for i := range idols {
			if idols[i].ID == id {
				deleted := idols[i]
				idols = append(idols[:i], idols[i+1:]...)
				writeJSON(w, http.StatusOK, deleted)
				return
			}
		}
		writeError(w, http.StatusNotFound, "idol not found")
	case http.MethodOptions:
		handleOptions(w, r)
	default:
		handleOptions(w, r)
	}
}

// handleUsers mengembalikan daftar user dari config.yaml (hanya username)
func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var out []map[string]string
	for _, u := range config.Users {
		out = append(out, map[string]string{"username": u.Username})
	}
	writeJSON(w, http.StatusOK, out)
}

// handleLogin menerima username/password, generate token jika valid
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if pass, ok := usersMap[in.Username]; !ok || pass != in.Password {
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	token, info := generateToken(in.Username)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_in": int(time.Until(info.ExpiresAt).Seconds()),
	})
}

// handleLogout menghapus token dari server
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	revokeToken(token)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logout success"})
}

// --- MAIN ---
func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Gagal load config: %v", err)
	}
	mux := http.NewServeMux()

	// API endpoint
	mux.HandleFunc("/api/idols", handleIdols)
	mux.HandleFunc("/api/idols/", handleIdolByID)
	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)

	// Serve static files (login.html, index.html)
	mux.Handle("/login.html", http.FileServer(http.Dir(".")))
	mux.Handle("/index.html", http.FileServer(http.Dir(".")))
	mux.Handle("/", http.FileServer(http.Dir("."))) // fallback ke static

	// Compose middleware: CORS -> Logging -> Auth -> Handler
	handler := corsMiddleware(loggingMiddleware(authMiddleware(mux)))

	addr := ":" + envPort
	log.Printf("Server running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
