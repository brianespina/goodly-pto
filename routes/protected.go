package routes

import (
	"fmt"
	"goodly-pto/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterProtectedRoutes(r *gin.RouterGroup, pool *pgxpool.Pool) {
	r.GET("/users", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "users.html", nil)
	})

	r.GET("/", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		user := new(models.User)
		var vacation_leave, sick_leave float64
		pool.QueryRow(ctx, `SELECT 
    u.name,
    u.email,
    COALESCE(MAX(CASE WHEN pt.id= 2 THEN pb.balance END), 0.0) AS vacation_leave,
    COALESCE(MAX(CASE WHEN pt.id = 1 THEN pb.balance END), 0.0) AS sick_leave
FROM users u
LEFT JOIN pto_balances pb ON u.id = pb.user_id
LEFT JOIN pto_types pt ON pb.pto_type_id = pt.id
WHERE u.id = $1
GROUP BY u.id, u.name, u.email
ORDER BY u.id;`, user_id).Scan(&user.Name, &user.Email, &vacation_leave, &sick_leave)
		// TODO: err hanling
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			"name":     user.Name,
			"email":    user.Email,
			"vacation": vacation_leave,
			"sick":     sick_leave,
		})
	})
	r.POST("/logout", func(ctx *gin.Context) {
		session_id, _ := ctx.Get("session_id")
		tag, err := pool.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session_id)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tag)
		ctx.SetCookie(
			"session_id", // cookie name
			"",           // empty value
			-1,           // maxAge: -1 means "delete now"
			"/",          // path
			"localhost",  // domain (must match how it was set)
			false,        // secure (true if using HTTPS)
			true,         // httpOnly
		)
		ctx.Redirect(http.StatusSeeOther, "/login")
	})

	r.GET("/submit-pto", func(ctx *gin.Context) {
		today := time.Now().Format("2006-01-02")
		ctx.HTML(http.StatusOK, "request-form.html", gin.H{
			"today": today,
		})
	})
	r.POST("/submit-pto", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		var request models.PTORequest
		var balance float64

		start_date := ctx.PostForm("start_date")
		end_date := ctx.PostForm("end_date")
		pto_type_raw := ctx.PostForm("type")
		pto_type, _ := strconv.Atoi(pto_type_raw)
		var pto_count float64

		err := pool.QueryRow(ctx, "SELECT count_weekdays($1, $2)", start_date, end_date).Scan(&pto_count)
		if err != nil {
			fmt.Println(err)
			return
		}

		request.StartDate = start_date
		request.EndDate = end_date
		request.UserID = user_id.(int)
		request.TypeID = pto_type
		request.Days = pto_count

		err = pool.QueryRow(ctx, "SELECT balance FROM pto_balances WHERE user_id = $1 AND pto_type_id = $2", request.UserID, request.TypeID).Scan(&balance)
		if err != nil {
			fmt.Println(err)
			return
		}
		if request.Days > balance {
			fmt.Println("Insuficient Balance")
			return
		}
		fmt.Println("Request Valid")

		tag, err := pool.Exec(ctx, "INSERT INTO pto_requests (user_id, pto_type_id, start_date, end_date, days) VALUES ($1, $2, $3, $4, $5)", request.UserID, request.TypeID, request.StartDate, request.EndDate, request.Days)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tag)
		ctx.IndentedJSON(http.StatusOK, request)
	})
	r.POST("/pto-requests/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		_, err := pool.Exec(ctx, "UPDATE pto_requests SET status = $1 WHERE id = $2", models.StatusApproved, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx.String(http.StatusOK, "approved")
	})
	r.GET("/pto-requests", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		var requests []models.PTORequest
		rows, err := pool.Query(ctx, `SELECT DISTINCT 
pr.id,
pt.title,
users.name as requester,
pr.days,
pr.status
FROM users 
JOIN pto_requests pr on pr.user_id = users.id
JOIN pto_types pt on pr.pto_type_id = pt.id
JOIN role_management rm on users.role_id = rm.managed_role_id
JOIN users mu on rm.manager_role_id = mu.role_id
WHERE mu.id = $1
`, user_id)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var request models.PTORequest
			err := rows.Scan(&request.Id, &request.Title, &request.User, &request.Days, &request.Status)
			requests = append(requests, request)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "Scan error: %v", err)
				return
			}
		}

		ctx.HTML(http.StatusOK, "pto-requests.html", gin.H{
			"requests": requests,
		})
	})

}
