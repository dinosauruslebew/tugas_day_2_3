package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
    
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if v != nil {
        _ = json.NewEncoder(w).Encode(v)
    }
}

func HandleSecretData(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"msg": "data rahasia"})
}

func HandleUsers(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
            return
        }
        rows, err := db.Query("SELECT id, username, role, created_at, updated_at, version FROM users WHERE deleted_at IS NULL ORDER BY id")
        if err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
            return
        }
        defer rows.Close()
        type user struct {
            ID        int64  `json:"id"`
            Username  string `json:"username"`
            Role      string `json:"role"`
            CreatedAt string `json:"created_at"`
            UpdatedAt string `json:"updated_at"`
            Version   int    `json:"version"`
        }
        var list []user
        for rows.Next() {
            var u user
            if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt, &u.Version); err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "scan error"})
                return
            }
            list = append(list, u)
        }
        writeJSON(w, http.StatusOK, list)
    }
}

func HandleIdols(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            rows, err := db.Query("SELECT id, name, \"group_name\", position FROM idols WHERE deleted_at IS NULL ORDER BY id")
            if err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
                return
            }
            defer rows.Close()
            type idol struct {
                ID       int64  `json:"id"`
                Name     string `json:"name"`
                Group    string `json:"group_name"`
                Position string `json:"position"`
            }
            var list []idol
            for rows.Next() {
                var it idol
                if err := rows.Scan(&it.ID, &it.Name, &it.Group, &it.Position); err != nil {
                    writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "scan error"})
                    return
                }
                list = append(list, it)
            }
            writeJSON(w, http.StatusOK, list)
        case http.MethodPost:
            var in struct {
                Name     string `json:"name"`
                Group    string `json:"group_name"`
                Position string `json:"position"`
            }
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
                writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
                return
            }
            res := db.QueryRow("INSERT INTO idols (name, \"group_name\", position) VALUES ($1,$2,$3) RETURNING id", in.Name, in.Group, in.Position)
            var id int64
            if err := res.Scan(&id); err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "insert error"})
                return
            }
            writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "name": in.Name, "group_name": in.Group, "position": in.Position})
        default:
            w.WriteHeader(http.StatusNoContent)
        }
    }
}

func HandleIdolByID(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
        if id == "" {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
            return
        }
        switch r.Method {
        case http.MethodPut:
            var in struct {
                Name     string `json:"name"`
                Group    string `json:"group_name"`
                Position string `json:"position"`
            }
            if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
                writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
                return
            }
            if _, err := db.Exec("UPDATE idols SET name=$1, \"group_name\"=$2, position=$3, updated_at=NOW(), version=version+1 WHERE id=$4 AND deleted_at IS NULL", in.Name, in.Group, in.Position, id); err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update error"})
                return
            }
            writeJSON(w, http.StatusOK, map[string]interface{}{"id": id, "name": in.Name, "group_name": in.Group, "position": in.Position})
        case http.MethodDelete:
            if _, err := db.Exec("UPDATE idols SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL", id); err != nil {
                writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete error"})
                return
            }
            writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
        default:
            w.WriteHeader(http.StatusNoContent)
        }
    }
}


