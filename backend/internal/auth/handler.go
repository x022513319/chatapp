package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	DB  *sql.DB
	RDB *redis.Client
}

type registerReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type loginReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !usernameOK(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username format"})
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "hash error"})
		return
	}

	var id int64
	err = h.DB.QueryRow(`
        INSERT INTO users (username, nickname, email, password_hash)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, req.Username, nullIfEmpty(req.Nickname), req.Email, hash).Scan(&id)

	if err != nil {
		var pqe *pq.Error
		if errors.As(err, &pqe) && pqe.Code == "23505" { // duplicate key value violates
			if strings.Contains(pqe.Constraint, "users_username_key") {
				c.JSON(409, gin.H{"error": "username already taken"})
				return
			}
			if strings.Contains(pqe.Constraint, "users_email_key") {
				c.JSON(409, gin.H{"error": "username already registered"})
				return
			}
		}
		c.JSON(500, gin.H{"error": "db error"})
	}
	c.JSON(201, gin.H{"id": id})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// --- read env var ---
	maxAttempts := atoiDefault(os.Getenv("LOGIN_MAX_ATTEMPTS"), 5)
	blockMinutes := atoiDefault(os.Getenv("LOGIN_BLOCK_MINUTES"), 15)
	windowSeconds := atoiDefault(os.Getenv("LOGIN_WINDOW_SECONDS"), 900)

	ctx := context.Background()

	keyAttempts := "login:attempts:" + strings.ToLower(req.Email)
	keyBlock := "login:block" + strings.ToLower(req.Email)

	if ttl, err := h.RDB.TTL(ctx, keyBlock).Result(); err == nil && ttl > 0 {
		c.JSON(429, gin.H{"error": "too many attempts, try later", "retry_after_seconds": int(ttl.Seconds())})
		return
	}

	var (
		id     int64
		hash   string
		status int
	)
	err := h.DB.QueryRow(`
        SELECT id, password_hash, status
        FROM users
        WHERE email = $1
    `, req.Email).Scan(&id, &hash, &status)

	checkFailed := func() {
		n, err := h.RDB.Incr(ctx, keyAttempts).Result()
		if err == nil && n == 1 {
			// attempt failed first time
			h.RDB.Expire(ctx, keyAttempts, time.Duration(windowSeconds)*time.Second)
		}

		if n >= int64(maxAttempts) {
			h.RDB.Set(ctx, keyBlock, 1, time.Duration(blockMinutes)*time.Minute)
		}
		c.JSON(401, gin.H{"error": "invalid email or password"})
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			checkFailed()
			return

		}
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	if status != 1 {
		c.JSON(403, gin.H{"error": "account disabled"})
		return
	}
	if !CheckPassword(hash, req.Password) {
		checkFailed()
		return
	}

	// login successed
	h.RDB.Del(ctx, keyAttempts)
	h.RDB.Del(ctx, keyBlock)

	// update the last_login_at field
	_, _ = h.DB.Exec(`UPDATE users SET last_login_at = $1 WHERE id = $2`, time.Now(), id)

	// issue JWT
	token, err := SignAccessToken(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "token error"})
		return
	}

	c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer"})

}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}

func (h *Handler) Me(c *gin.Context) {
	uid := c.GetInt64("userID")
	var (
		username, nickname, email string
		createdAt                 time.Time
	)

	err := h.DB.QueryRow(`
        SELECT username, COALESCE(nickname, ''), email, created_at
        FROM users WHERE id=$1
    `, uid).Scan(&username, &nickname, &email, &createdAt)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	c.JSON(200, gin.H{
		"id":         uid,
		"username":   username,
		"nickname":   nickname,
		"email":      email,
		"created_at": createdAt,
	})
}

// --- helpers ---
func usernameOK(s string) bool {
	if len(s) < 3 || len(s) > 32 {
		return false
	}

	for i, r := range s {
		if i == 0 && !isLetter(r) {
			return false
		}
		if !isLetter(r) && !isDigit(r) && r != '.' && r != '_' && r != '-' {
			return false
		}
	}
	return true
}
func isLetter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
func isDigit(r rune) bool  { return r >= '0' && r <= '9' }
func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
