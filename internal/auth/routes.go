package auth

import (
	"net/http"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RegisterRoutes(r *gin.Engine, pool *pgxpool.Pool) {
	r.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", nil)
	})

	r.GET("/register/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		var name string
		err := pool.QueryRow(ctx, "SELECT name FROM users WHERE Id=$1", id).Scan(&name)
		if err != nil {
			fmt.Printf("Error retrieving user in register\nDatabase error: %v", err)
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
			fmt.Printf("Error hashing password\nError: %v", err)
			return
		}
		_, err = pool.Exec(ctx, "UPDATE users SET password = $1 WHERE id = $2", hash, id)
		if err != nil {
			fmt.Printf("Error resetting password\nDatabase error: %v", err)
			return
		}
		ctx.Redirect(http.StatusFound, "/login")
	})
	r.POST("/login", func(ctx *gin.Context) {
		stdCtx := ctx.Request.Context()
		var email string
		var hashed_password *string
		var id int

		form_email := ctx.PostForm("email")
		form_password := ctx.PostForm("password")

		err := pool.QueryRow(stdCtx, "SELECT email, password, id FROM users WHERE email = $1", form_email).Scan(&email, &hashed_password, &id)
		if err != nil {
			// TODO: Handle wrong email
			fmt.Printf("Wrong/No email\nDatabase error: %v", err)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(*hashed_password), []byte(form_password))
		if err != nil {
			// TODO: Handle wrong password
			fmt.Printf("Wrong Password\nError: %v", err)
			return
		}

		session_id := uuid.New().String()
		expires := time.Now().Add(24 * time.Hour)

		_, err = pool.Exec(ctx, "INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)", session_id, id, expires)
		if err != nil {
			fmt.Printf("Error inserting into sessions\nDatabase error: %v", err)
			return
		}
		ctx.SetCookie("session_id", session_id, 86400, "/", "localhost", false, true)
		ctx.Redirect(http.StatusSeeOther, "/")
	})

	r.POST("/logout", func(ctx *gin.Context) {
		session_id, _ := ctx.Get("session_id")
		_, err := pool.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session_id)
		if err != nil {
			fmt.Printf("Error logging out\nDatabase error: %v", err)
			return
		}
		ctx.SetCookie(
			"session_id",
			"",
			-1,
			"/",
			"localhost",
			false,
			true,
		)
		ctx.Redirect(http.StatusSeeOther, "/login")
	})
}
