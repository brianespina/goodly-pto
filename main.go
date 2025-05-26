package main

import (
	"context"
	"fmt"
	"os"

	"goodly-pto/internal/auth"
	"goodly-pto/internal/pto"
	"goodly-pto/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	pool, err := pgxpool.New(context.Background(), os.Getenv("DBSTRING"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	r := gin.Default()
	r.Static("/static", "./static")
	r.SetTrustedProxies(nil)
	r.LoadHTMLGlob("templates/*")

	authGroup := r.Group("/")
	authGroup.Use(auth.AuthRequired(pool))

	pto.RegisterRoutes(authGroup, pool)
	user.RegisterRoutes(authGroup, pool)

	auth.RegisterRoutes(r, pool)

	r.Run()
}
