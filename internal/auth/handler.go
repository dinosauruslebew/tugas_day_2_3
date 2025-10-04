package auth

import (
	"encoding/json"
	"net/http"
	"time"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *AuthService) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in loginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	username := in.Username
	// Validate only against YAML users; fallback to BASIC if YAML empty
	matched := false
	role := "user"
	if len(a.cfg.Users) > 0 {
		for _, u := range a.cfg.Users {
			if u.Username == username && u.Password == in.Password {
				matched = true
				if username == "admin" {
					role = "admin"
				}
				break
			}
		}
	} else {
		if username == a.cfg.Basic.Username && in.Password == a.cfg.Basic.Password {
			matched = true
			role = "admin"
		}
	}
	if !matched {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	token, exp, err := a.CreateToken(username, role)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token":      token,
		"expires_in": int(time.Until(exp).Seconds()),
	})
}

func (a *AuthService) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	authz := r.Header.Get("Authorization")
	if len(authz) < 8 {
		http.Error(w, "missing bearer token", http.StatusUnauthorized)
		return
	}
	token := authz[7:]
	claims, err := a.ParseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	a.Blacklist(token, claims.RegisteredClaims.ExpiresAt.Time)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "logout success"})
}
