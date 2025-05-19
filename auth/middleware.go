package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

func AuthRequired(db *pgxpool.Pool) gin.HandlerFunc {
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
		ctx.Set("session_id", uuid_session_id)
		ctx.Next()
	}
}
