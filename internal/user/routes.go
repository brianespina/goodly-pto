package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RenderTemplateWithPermission(ctx *gin.Context, code int, template string, data gin.H) {
	permission, ok := ctx.Get("permission")
	if ok {
		data["permission"] = permission
	}
	ctx.HTML(code, template, data)
}

func RegisterRoutes(r *gin.RouterGroup, pool *pgxpool.Pool) {
	r.GET("/", func(ctx *gin.Context) {
		user := new(User)
		var vacation_leave, sick_leave float64

		user_id, _ := ctx.Get("user_id")

		err := pool.QueryRow(ctx, `
			SELECT 
			u.name, 
			u.email,
			COALESCE(MAX(CASE WHEN pt.id= 2 THEN pb.balance END), 0.0) AS vacation_leave,
			COALESCE(MAX(CASE WHEN pt.id = 1 THEN pb.balance END), 0.0) AS sick_leave
			FROM users u
			LEFT JOIN pto_balances pb ON u.id = pb.user_id
			LEFT JOIN pto_types pt ON pb.pto_type_id = pt.id
			WHERE u.id = $1
			GROUP BY u.id, u.name, u.email
			ORDER BY u.id;
		`, user_id).Scan(&user.Name, &user.Email, &vacation_leave, &sick_leave)

		if err != nil {
			fmt.Printf("Error retreiving user in Dashboard\nDatabase error: %v", err)
			return
		}

		RenderTemplateWithPermission(ctx, http.StatusOK, "index.html", gin.H{
			"name":     user.Name,
			"email":    user.Email,
			"vacation": vacation_leave,
			"sick":     sick_leave,
		})
	})
}
