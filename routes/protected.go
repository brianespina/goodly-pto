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
		pool.QueryRow(ctx, `
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
		var balance float64
		var pto_count float64
		start_date := ctx.PostForm("start_date")
		end_date := ctx.PostForm("end_date")
		pto_type_raw := ctx.PostForm("type")
		pto_type, _ := strconv.Atoi(pto_type_raw)
		reason := ctx.PostForm("reason")
		err := pool.QueryRow(ctx, "SELECT count_weekdays($1, $2)", start_date, end_date).Scan(&pto_count)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = pool.QueryRow(ctx, "SELECT balance FROM pto_balances WHERE user_id = $1 AND pto_type_id = $2", user_id, pto_type).Scan(&balance)
		if err != nil {
			fmt.Println(err)
			return
		}
		if pto_count > balance {
			fmt.Println("Insuficient Balance")
			return
		}

		fmt.Println("Request Valid")

		tag, err := pool.Exec(
			ctx,
			"INSERT INTO pto_requests (user_id, pto_type_id, start_date, end_date, days, reason) VALUES ($1, $2, $3, $4, $5, $6)",
			user_id,
			pto_type,
			start_date,
			end_date,
			pto_count,
			reason,
		)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(tag)
		fmt.Println("request sent")
	})
	r.POST("/team-requests/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		_, err := pool.Exec(ctx, "UPDATE pto_requests SET status = $1 WHERE id = $2", models.StatusApproved, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx.String(http.StatusOK, "approved")
	})
	r.GET("/team-requests", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		var requests []models.PTORequest
		rows, err := pool.Query(ctx, `
			SELECT DISTINCT 
			pr.id,
			pt.title,
			users.name as requester,
			pr.days,
			pr.status,
			pr.reason
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
			err := rows.Scan(&request.Id, &request.Type, &request.User, &request.Days, &request.Status, &request.Reason)
			requests = append(requests, request)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "/team-requests", err)
				return
			}
		}

		ctx.HTML(http.StatusOK, "team-requests.html", gin.H{
			"requests": requests,
		})
	})
	r.GET("/my-requests", func(ctx *gin.Context) {

		user_id, _ := ctx.Get("user_id")
		var requests []models.PTORequest
		rows, err := pool.Query(ctx, `
			SELECT u.id, pt.title, u.name, pr.days, pr.status, pr.reason
			FROM pto_requests as pr
			JOIN pto_types pt on pt.id = pr.pto_type_id
			JOIN users u on u.id = pr.user_id
			WHERE u.id = $1
		`, user_id)

		if err != nil {
			fmt.Println(err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			var request models.PTORequest
			err := rows.Scan(&request.Id, &request.Type, &request.User, &request.Days, &request.Status, &request.Reason)
			requests = append(requests, request)
			if err != nil {
				ctx.String(http.StatusInternalServerError, "/team-requests", err)
				return
			}
		}
		ctx.HTML(http.StatusOK, "my-requests.html", gin.H{
			"requests": requests,
		})
	})

}
