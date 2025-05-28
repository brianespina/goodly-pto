package pto

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PTOService struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *PTOService {
	return &PTOService{db: db}
}

func (s *PTOService) MyRequests(ctx *gin.Context, user_id any) ([]PTORequest, error) {
	var requests []PTORequest
	rows, err := s.db.Query(ctx, `
			SELECT u.id, pt.title, u.name, pr.days, pr.status, pr.reason, pr.start_date, pr.end_date
			FROM pto_requests as pr
			JOIN pto_types pt on pt.id = pr.pto_type_id
			JOIN users u on u.id = pr.user_id
			WHERE u.id = $1
		`, user_id)

	if err != nil {
		fmt.Printf("Error retrieving my requests\nDatabase error: %v", err)
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var request PTORequest
		err := rows.Scan(&request.Id, &request.Type, &request.User, &request.Days, &request.Status, &request.Reason, &request.StartDate, &request.EndDate)
		requests = append(requests, request)
		if err != nil {
			fmt.Printf("Error scanning my requests\nDatabase error: %v", err)
			return nil, err
		}
	}
	return requests, nil
}
