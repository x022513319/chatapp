package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func setup(t *testing.T) (*Handler, *gin.Engine, sqlmock.Sqlmock, func()) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// fake DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}

	// fake redis
	mr, _ := miniredis.Run()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	h := &Handler{DB: db, RDB: rdb}

	r := gin.New()
	r.POST("/api/register", h.Register)
	r.POST("/api/login", h.Login)
	r.GET("/api/me", func(c *gin.Context) {
		c.Set("userID", int64(1))
		h.Me(c)
	})

	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("ACCESS_TOKEN_TTL_MIN", "5")

	cleanup := func() {
		_ = db.Close()
		mr.Close()
	}

	return h, r, mock, cleanup
}

func bodyJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestRegister_Success(t *testing.T) {
	_, r, mock, cleanup := setup(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", "Alice", "alice@example.com", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	req := httptest.NewRequest(http.MethodPost, "/api/register",
		strings.NewReader(bodyJSON(map[string]string{
			"username": "alice", "nickname": "Alice",
			"email": "alice@example.com", "password": "p@ssw0rd123",
		})))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRegister_ConflictEmail(t *testing.T) {
	_, r, mock, cleanup := setup(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnError(&pq.Error{
			Code:       "23505",
			Constraint: "users_email_key",
		})

	req := httptest.NewRequest(http.MethodPost, "/api/register",
		strings.NewReader(bodyJSON(map[string]string{
			"username": "bob", "nickname": "Bob",
			"email": "bob@example.com", "password": "whatever",
		})))
	req.Header.Set("content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestLogin_WrongPassword_IncrementAttempts(t *testing.T) {
	h, r, mock, cleanup := setup(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, password_hash, status FROM users WHERE email = \$1`).
		WithArgs("wrong@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "password_hash", "status"}).
			AddRow(1, "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1))

	req := httptest.NewRequest(http.MethodPost, "/api/login",
		strings.NewReader(bodyJSON(map[string]string{
			"email": "wrong@example.com", "password": "not-correct",
		})))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", w.Code, w.Body.String())
	}

	ctx := context.Background()
	n, err := h.RDB.Get(ctx, "login:attempts:wrong@example.com").Int()
	if err != nil {
		t.Fatalf("expected attempts key exists, got err: %v", err)
	}
	if n <= 0 {
		t.Fatalf("expected attempts > 0, got %d", n)
	}

	ttl, err := h.RDB.TTL(ctx, "login:attempts:wrong@example.com").Result()
	if err != nil {
		t.Fatalf("TTL error: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("expected positive TTL, got %v", ttl)
	}
}

func TestLogin_Blocked(t *testing.T) {
	h, r, mock, cleanup := setup(t)
	defer cleanup()

	ctx := context.Background()
	h.RDB.Set(ctx, "login:block:alice@example.com", "1", 2*time.Minute)

	req := httptest.NewRequest(http.MethodPost, "/api/login",
		strings.NewReader(bodyJSON(map[string]string{
			"email": "alice@example.com", "password": "x",
		})))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d body=%s", w.Code, w.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestMe_Success(t *testing.T) {
	_, r, mock, cleanup := setup(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT username, COALESCE\(nickname, ''\), email, created_at FROM users WHERE id=\$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"username", "nickname", "email", "created_at"}).
			AddRow("alice", "Alice", "alice@example.com", time.Now()))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
}
