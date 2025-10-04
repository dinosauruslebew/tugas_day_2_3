package auth

import (
    "database/sql"
    "errors"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "kpopapi/config"
)

type AuthService struct {
    db       *sql.DB
    cfg      config.AppConfig
    jwtKey   []byte
    blacklist struct {
        sync.RWMutex
        m map[string]time.Time
    }
}

func NewAuthService(db *sql.DB, cfg config.AppConfig) *AuthService {
    as := &AuthService{db: db, cfg: cfg, jwtKey: []byte("secret_dev_key_change_me")}
    as.blacklist.m = make(map[string]time.Time)
    return as
}

type Claims struct {
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// CreateToken returns a signed JWT for the given username/role
func (a *AuthService) CreateToken(username, role string) (string, time.Time, error) {
    expiresAt := time.Now().Add(1 * time.Hour)
    claims := &Claims{
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiresAt),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(a.jwtKey)
    return signed, expiresAt, err
}

func (a *AuthService) ParseToken(tokenStr string) (*Claims, error) {
    if a.isBlacklisted(tokenStr) {
        return nil, errors.New("token blacklisted")
    }
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return a.jwtKey, nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}

func (a *AuthService) Blacklist(token string, exp time.Time) {
    a.blacklist.Lock()
    a.blacklist.m[token] = exp
    a.blacklist.Unlock()
}

func (a *AuthService) isBlacklisted(token string) bool {
    a.blacklist.RLock()
    exp, ok := a.blacklist.m[token]
    a.blacklist.RUnlock()
    if !ok {
        return false
    }
    return time.Now().Before(exp)
}

// JWTMiddleware enforces Authorization: Bearer <token> on all routes except login and swagger/static
func JWTMiddleware(auth *AuthService, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        if strings.HasPrefix(path, "/api/login") ||
            strings.HasPrefix(path, "/swagger") ||
            path == "/" ||
            strings.HasSuffix(path, ".html") ||
            strings.HasSuffix(path, ".js") ||
            strings.HasSuffix(path, ".css") {
            next.ServeHTTP(w, r)
            return
        }
        authz := r.Header.Get("Authorization")
        if !strings.HasPrefix(authz, "Bearer ") {
            http.Error(w, "missing bearer token", http.StatusUnauthorized)
            return
        }
        token := strings.TrimPrefix(authz, "Bearer ")
        if _, err := auth.ParseToken(token); err != nil {
            http.Error(w, "invalid or expired token", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}


