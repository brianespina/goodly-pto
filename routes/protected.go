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

const (
	FieldStartDateRequired string = "Start date Required"
	FieldEndDateRequired   string = "End date Required"
)

func RenderTemplateWithPermission(ctx *gin.Context, code int, template string, data gin.H) {
	permission, ok := ctx.Get("permission")
	if ok {
		data["permission"] = permission
	}
	ctx.HTML(code, template, data)
}

func RegisterProtectedRoutes(r *gin.RouterGroup, pool *pgxpool.Pool) {

	r.GET("/", func(ctx *gin.Context) {
		user := new(models.User)
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
	r.POST("/logout", func(ctx *gin.Context) {
		session_id, _ := ctx.Get("session_id")
		_, err := pool.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session_id)
		if err != nil {
			fmt.Printf("Error logging out\nDatabase error: %v", err)
			return
		}
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
		RenderTemplateWithPermission(ctx, http.StatusOK, "request-form.html", gin.H{
			"today": today,
		})
	})
	r.POST("/submit-pto", func(ctx *gin.Context) {
		var balance float64
		var pto_count float64

		var validationMsgs []string

		user_id, _ := ctx.Get("user_id")
		start_date := ctx.PostForm("start_date")
		end_date := ctx.PostForm("end_date")
		pto_type_raw := ctx.PostForm("type")
		pto_type, err := strconv.Atoi(pto_type_raw)
		reason := ctx.PostForm("reason")

		//Start Date Validation
		if start_date == "" {
			validationMsgs = append(validationMsgs, FieldStartDateRequired)
		}
		today := time.Now().Truncate(24 * time.Hour)
		start_date_parsed, err := time.Parse("2006-01-02", start_date)
		if start_date_parsed.Before(today) {
			validationMsgs = append(validationMsgs, "Start date can't be before today")
		}
		if err != nil {
			validationMsgs = append(validationMsgs, "Start date invalid")
		}

		//End Date Validation
		if end_date == "" {
			validationMsgs = append(validationMsgs, FieldEndDateRequired)
		}
		_, err = time.Parse("2006-01-02", start_date)
		if err != nil {
			validationMsgs = append(validationMsgs, "End date invalid")
		}
		if reason == "" {
			validationMsgs = append(validationMsgs, "Reason is required")
		}

		if len(validationMsgs) > 0 {
			// TODO: send error msgs
			ctx.Redirect(http.StatusSeeOther, "/submit-pto")
			fmt.Println(validationMsgs)
			return
		}

		if err != nil {
			fmt.Printf("Error converting string in pto submission\nError: %v", err)
			return
		}

		err = pool.QueryRow(ctx, "SELECT count_weekdays($1, $2)", start_date, end_date).Scan(&pto_count)
		if err != nil {
			fmt.Printf("Error counting weekdays\nDatabase error: %v", err)
			return
		}

		err = pool.QueryRow(ctx, "SELECT balance FROM pto_balances WHERE user_id = $1 AND pto_type_id = $2", user_id, pto_type).Scan(&balance)
		if err != nil {
			fmt.Printf("Error retrieving pto balance\nDatabase error: %v", err)
			return
		}
		if pto_count > balance {
			// TODO: validation
			return
		}
		_, err = pool.Exec(
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
			fmt.Printf("Error submitting request\nDatabase error: %v", err)
			return
		}
		ctx.Redirect(http.StatusSeeOther, "/submit-pto")
	})
	r.POST("/team-requests/:id", func(ctx *gin.Context) {
		request_id := ctx.Param("id")
		var days, requester_id, pto_type int
		if err := pool.QueryRow(ctx, "SELECT days, user_id, pto_type_id FROM pto_requests WHERE id = $1", request_id).Scan(&days, &requester_id, &pto_type); err != nil {
			fmt.Printf("Error request does not exist\nDatabase error: %v", err)
			return
		}
		if _, err := pool.Exec(ctx, "UPDATE pto_balances SET balance = balance - $1 WHERE user_id = $2 AND pto_type_id = $3", days, requester_id, pto_type); err != nil {
			fmt.Printf("Error cant update PTO balance\nDatabase error: %v", err)
			return
		}
		_, err := pool.Exec(ctx, "UPDATE pto_requests SET status = $1 WHERE id = $2", models.StatusApproved, request_id)
		if err != nil {
			fmt.Printf("Error approving request\nDatabase error: %v", err)
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
			fmt.Printf("Error retrieving team requests\nDatabase error: %v", err)
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
		RenderTemplateWithPermission(ctx, http.StatusOK, "team-requests.html", gin.H{
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
			fmt.Printf("Error retrieving my requests\nDatabase error: %v", err)
			return
		}

		defer rows.Close()
		for rows.Next() {
			var request models.PTORequest
			err := rows.Scan(&request.Id, &request.Type, &request.User, &request.Days, &request.Status, &request.Reason)
			requests = append(requests, request)
			if err != nil {
				fmt.Printf("Error scanning my requests\nDatabase error: %v", err)
				return
			}
		}
		RenderTemplateWithPermission(ctx, http.StatusOK, "my-requests.html", gin.H{
			"requests": requests,
		})
	})

}
