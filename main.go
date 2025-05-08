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
	r.LoadHTMLGlob("templates/*")
	r.Static("/js", "./js")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"title": os.Getenv("TITLE"),
		})
	})

	r.POST("/test", func(ctx *gin.Context) {
		var roles []Role
		rows, err := conn.Query(ctx, "select title from roles")
		if err != nil {
			//handle errors
		}
		defer rows.Close()

		for rows.Next() {
			var role Role
			if err := rows.Scan(&role.Title); err != nil {
				ctx.String(http.StatusInternalServerError, "Row Scan Error %v", err)
				return
			}
			roles = append(roles, role)
		}

		if err := rows.Err(); err != nil {
			ctx.String(http.StatusInternalServerError, "Rows Iteration Error %v", err)
			return
		}

		ctx.HTML(http.StatusOK, "list.html", gin.H{
			"roles": roles,
		})
	})

	r.Run()
}
