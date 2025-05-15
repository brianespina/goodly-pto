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
	Name  string
	Email string
	Role  string
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

	r.GET("/users", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "users.html", nil)
	})
	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/db", func(ctx *gin.Context) {
		rows, err := conn.Query(ctx, "select title from roles")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rows.Close()

		var roles []Role

		for rows.Next() {
			var role Role
			if err := rows.Scan(&role.Title); err != nil {
				fmt.Println(err)
				return
			}
			roles = append(roles, role)
		}
		if err := rows.Err(); err != nil {
			fmt.Println(err)
			return
		}
		ctx.HTML(http.StatusOK, "roles.html", gin.H{
			"roles": roles,
		})
	})

	r.Run()
}
