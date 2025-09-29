package main

import (
	apihttp "backend/internal/api"
	"backend/internal/auth"
	authpkg "backend/internal/auth"
	"backend/internal/store"
	"backend/internal/ws"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func mustOpenDB() *sql.DB {
	dsn := os.Getenv("POSTGRES_URL") // data source name
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func mustOpenRedis() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	dbIdx := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			dbIdx = n
		}
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: dbIdx})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis connect error:", err)
	}
	return rdb
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := authpkg.ParseAndValidate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// get userID (sub)
		uid, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "bad sub"})
			return
		}
		c.Set("userID", uid)
		c.Next()
	}
}

func wsAuthMiddleWare() gin.HandlerFunc {

	return func(c *gin.Context) {
		var token string

		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimPrefix(h, "Bearer ")
		} else if q := c.Query("token"); q != "" {
			token = q
		} else if ck, err := c.Cookie("access_token"); err == nil {
			token = ck
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "mission token", "code": "UNAUTHORIZED"})
			return
		}
		claim, err := auth.ParseAndValidate(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token", "code": "UNAUTHORIZED"})
			return
		}
		uid, err := strconv.ParseInt(claim.Subject, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "bad sub", "code": "UNAUTHORIZED"})
			return
		}
		c.Set("userID", uid)
		c.Next()
	}
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, fallback to system env")
	}
}

func main() {
	// check environment variables
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET not set")
	}

	db := mustOpenDB()
	defer db.Close()

	rdb := mustOpenRedis()
	defer rdb.Close()

	authH := &authpkg.Handler{DB: db, RDB: rdb}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Auth APIs
	apiGroup := r.Group("/api")
	apiGroup.POST("/register", authH.Register)
	apiGroup.POST("/login", authH.Login)
	apiGroup.GET("/me", authMiddleware(), authH.Me)

	st := store.NewFormDB(db)

	// Chat REST
	chatAPI := &apihttp.Handler{Store: st}
	apihttp.Mount(apiGroup, chatAPI, authMiddleware())

	// WebSocket
	wsh := &ws.Handler{Store: st, Hub: ws.NewHub()}
	r.GET("/ws", wsAuthMiddleWare(), wsh.Handle)

	log.Println("backend running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
