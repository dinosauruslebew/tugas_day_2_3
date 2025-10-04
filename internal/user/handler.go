package user

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
)

type Idol struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Group    string `json:"group"`
    Position string `json:"position"`
}

type Handler struct {
    db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
    return &Handler{db: db}
}

// GET /api/idols
func (h *Handler) GetIdols(w http.ResponseWriter, r *http.Request) {
    rows, err := h.db.Query(`SELECT id, name, "group", position FROM idols`)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var idols []Idol
    for rows.Next() {
        var i Idol
        if err := rows.Scan(&i.ID, &i.Name, &i.Group, &i.Position); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        idols = append(idols, i)
    }
    json.NewEncoder(w).Encode(idols)
}

// POST /api/idols
func (h *Handler) AddIdol(w http.ResponseWriter, r *http.Request) {
    var i Idol
    if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    _, err := h.db.Exec(`INSERT INTO idols (name, "group", position) VALUES ($1, $2, $3)`, i.Name, i.Group, i.Position)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
}

// PUT /api/idols/{id}
func (h *Handler) UpdateIdol(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    id, _ := strconv.Atoi(parts[len(parts)-1])

    var i Idol
    if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := h.db.Exec(`UPDATE idols SET name=$1, "group"=$2, position=$3 WHERE id=$4`, i.Name, i.Group, i.Position, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

// DELETE /api/idols/{id}
func (h *Handler) DeleteIdol(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    id, _ := strconv.Atoi(parts[len(parts)-1])

    _, err := h.db.Exec(`DELETE FROM idols WHERE id=$1`, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
}

// POST /api/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    // hapus token dari server memory (kalau pakai map/redis dsb)
    // sekarang minimal kasih response sukses
    json.NewEncoder(w).Encode(map[string]string{"message": "logout success"})
}
