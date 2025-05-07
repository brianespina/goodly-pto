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

func main() {

	users := []User{
		{
			Name:  "Brian Espina",
			Email: "brian@goodlygrowth.com",
			Role:  "Senior Web Developer",
		},
		{
			Name:  "Kenneth Romero",
			Email: "kenneth@goodlygrowth.com",
			Role:  "SEO L3",
		},
	}

	godotenv.Load()
	conn, err := pgx.Connect(context.Background(), os.Getenv("DBSTRING"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.LoadHTMLGlob("templates/*")
	r.Static("/js", "./js")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": os.Getenv("TITLE"),
		})
	})

	r.POST("/test", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "list.tmpl", gin.H{
			"users": users,
		})
	})

	r.Run()
}
