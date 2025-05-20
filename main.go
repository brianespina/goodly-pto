package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"goodly-pto/auth"
	"goodly-pto/routes"
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
	r.SetTrustedProxies(nil)
	r.Static("/js", "./js")
	r.LoadHTMLGlob("templates/*")

	authGroup := r.Group("/")
	authGroup.Use(auth.AuthRequired(pool))

	routes.RegisterProtectedRoutes(authGroup, pool)
	routes.RegisterRoutes(r, pool)

	r.Run()
}
