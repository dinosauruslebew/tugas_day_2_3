package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "strings"
)

// Idol merepresentasikan satu idol K-Pop
type Idol struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Group    string `json:"group"`
    Position string `json:"position"`
}

// in-memory storage untuk idols dan auto-increment id
var (
    idols  = []Idol{
        {ID: "1", Name: "Jisung", Group: "NCT", Position: "Main Dancer"},
        {ID: "2", Name: "Karina", Group: "AESPA", Position: "Leader"},
        {ID: "3", Name: "Ahyeon", Group: "BABYMONSTER", Position: "Main Vocal"},
    }
    nextID = 4
)

// addCORS menambahkan header CORS ke response
func addCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-Control-Max-Age", "86400")
}

// writeJSON membantu menuliskan response JSON dengan header yang sesuai
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
    addCORS(w)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if data == nil {
        return
    }
    _ = json.NewEncoder(w).Encode(data)
}

// writeError menuliskan error dalam format JSON
func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}

// loggingMiddleware mencatat method dan URL tiap request
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

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
    // Path yang diharapkan: /api/idols/{id}
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

// main men-setup routes dan menjalankan HTTP server pada port 8080
func main() {
    mux := http.NewServeMux()

    // Route untuk collection
    mux.HandleFunc("/api/idols", func(w http.ResponseWriter, r *http.Request) {
        addCORS(w)
        if r.Method == http.MethodOptions {
            handleOptions(w, r)
            return
        }
        if r.URL.Path != "/api/idols" {
            writeError(w, http.StatusNotFound, "not found")
            return
        }
        handleIdols(w, r)
    })

    // Route untuk item by id
    mux.HandleFunc("/api/idols/", func(w http.ResponseWriter, r *http.Request) {
        addCORS(w)
        if r.Method == http.MethodOptions {
            handleOptions(w, r)
            return
        }
        handleIdolByID(w, r)
    })

    // Bungkus dengan logging middleware
    handler := loggingMiddleware(mux)

    // Jalankan server
    addr := ":8080"
    log.Printf("Server running on http://localhost%s", addr)
    if err := http.ListenAndServe(addr, handler); err != nil {
        log.Fatal(err)
    }
}
