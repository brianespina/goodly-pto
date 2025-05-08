package main

import (
	"context"
	"net/http"

	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type User struct {
	Name  string `json: name`
	Email string `json: email`
	Role  string `json: role`
}
type Role struct {
	Title string
}

func main() {

	godotenv.Load()
	conn, err := pgx.Connect(context.Background(), os.Getenv("DBSTRING"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Static("/js", "./js")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "base.html", gin.H{
			"title": "Goodly PTO",
		})
	})
	r.GET("/roles", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "roles.html", gin.H{})
	})
	r.GET("/users", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "users.html", gin.H{})
	})

	r.Run()
}
