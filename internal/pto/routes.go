package pto

import (
	"fmt"
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

func RegisterRoutes(r *gin.RouterGroup, pool *pgxpool.Pool, service *PTOService) {
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
		_, err := pool.Exec(ctx, "UPDATE pto_requests SET status = $1 WHERE id = $2", StatusApproved, request_id)
		if err != nil {
			fmt.Printf("Error approving request\nDatabase error: %v", err)
			return
		}
		ctx.String(http.StatusOK, "approved")
	})
	r.GET("/team-requests", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		requests, err := service.TeamRequests(ctx, user_id)
		if err != nil {
			fmt.Printf("Error fetching team requests\n")
			return
		}
		RenderTemplateWithPermission(ctx, http.StatusOK, "team-requests.html", gin.H{
			"requests": requests,
		})
	})
	r.GET("/my-requests", func(ctx *gin.Context) {

		user_id, _ := ctx.Get("user_id")
		requests, err := service.MyRequests(ctx, user_id)
		if err != nil {
			fmt.Printf("Error fetching My requests\n")
			return
		}
		RenderTemplateWithPermission(ctx, http.StatusOK, "my-requests.html", gin.H{
			"requests": requests,
		})
	})

	r.POST("/my-requests", func(ctx *gin.Context) {
		user_id, _ := ctx.Get("user_id")
		requests, err := service.MyRequests(ctx, user_id)
		if err != nil {
			fmt.Printf("Error fetching My requests\n")
			return
		}
		ctx.HTML(http.StatusOK, "component-pto-list", gin.H{
			"requests": requests,
		})
	})
}
