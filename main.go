package main

import (
	"context"
	"net/http"
	"time"

	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name  string
	Email string
	Role  string
}
type Role struct {
	Title string
}

func AuthRequired(db *pgx.Conn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session_id, err := ctx.Cookie("session_id")
		var user_id int
		if err != nil {
			fmt.Println("no session ID in cookies")
			ctx.Redirect(http.StatusSeeOther, "/login")
			ctx.Abort()
			return
		}
		uuid_session_id, err := uuid.Parse(session_id)
		err = db.QueryRow(ctx, "SELECT user_id FROM sessions WHERE id = $1 AND expires_at > NOW()", uuid_session_id).Scan(&user_id)
		if err != nil {
			fmt.Println("no session in DB")
			fmt.Println(err)
			ctx.Redirect(http.StatusSeeOther, "/login")
			ctx.Abort()
			return
		}
		ctx.Set("user_id", user_id)
		ctx.Next()
	}
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

	auth := r.Group("/")
	auth.Use(AuthRequired(conn))
	auth.GET("/users", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "users.html", nil)
	})
	auth.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/register/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		var name string
		err := conn.QueryRow(ctx, "SELECT name FROM users WHERE Id=$1", id).Scan(&name)

		if err != nil {
			fmt.Println(err)
			return
		}
		ctx.HTML(http.StatusOK, "register.html", gin.H{
			"name": name,
		})
	})

	r.POST("/register/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		password := ctx.PostForm("password")
		passwordcopy := ctx.PostForm("passwordcopy")
		if password != passwordcopy {
			ctx.HTML(http.StatusOK, "register.html", gin.H{
				"flash": "Passwords did not match",
			})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			fmt.Println(err)
			return
		}
		commTag, err := conn.Exec(ctx, "UPDATE users SET password = $1 WHERE id = $2", hash, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(commTag)
		ctx.Redirect(http.StatusFound, "/login")
	})
	r.POST("/login", func(ctx *gin.Context) {
		stdCtx := ctx.Request.Context()
		var email string
		var hashed_password *string
		var id int

		form_email := ctx.PostForm("email")
		form_password := ctx.PostForm("password")

		err := conn.QueryRow(stdCtx, "SELECT email, password, id FROM users WHERE email = $1", form_email).Scan(&email, &hashed_password, &id)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(*hashed_password), []byte(form_password))
		if err != nil {
			// TODO: Handle wrong password
			fmt.Println(err)
			return
		}

		// TODO: Login logic here
		session_id := uuid.New().String()
		expires := time.Now().Add(24 * time.Hour)

		_, err = conn.Exec(ctx, "INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", session_id, id, expires)
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx.SetCookie("session_id", session_id, 86400, "/", "localhost", false, true)

		fmt.Println("Log me in")

	})
	r.GET("/db", func(ctx *gin.Context) {
		stdCtx := ctx.Request.Context()

		rows, err := conn.Query(stdCtx, "select title from roles")
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
		ctx.IndentedJSON(http.StatusOK, roles)
	})

	r.Run()
}
