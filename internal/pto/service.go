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

func (s *PTOService) GetMyRequests(ctx *gin.Context, user_id any) ([]PTORequest, error) {
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

func (s *PTOService) GetTeamRequests(ctx *gin.Context, user_id any) ([]PTORequest, error) {
	var requests []PTORequest
	rows, err := s.db.Query(ctx, `
			SELECT DISTINCT 
			pr.id,
			pt.title,
			pr.start_date,
			pr.end_date,
			users.name as requester,
			pr.days,
			pr.status,
			pr.reason,
			pr.created_at
			FROM users 
			JOIN pto_requests pr on pr.user_id = users.id
			JOIN pto_types pt on pr.pto_type_id = pt.id
			JOIN role_management rm on users.role_id = rm.managed_role_id
			JOIN users mu on rm.manager_role_id = mu.role_id
			WHERE mu.id = $1
			AND pr.status = 'pending'
			ORDER BY pr.created_at DESC
		`, user_id)
	if err != nil {
		fmt.Printf("Error retrieving team requests\nDatabase error: %v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var request PTORequest
		err := rows.Scan(&request.Id, &request.Type, &request.StartDate, &request.EndDate, &request.User, &request.Days, &request.Status, &request.Reason, &request.CreatedDate)
		requests = append(requests, request)
		if err != nil {
			return nil, err
		}
	}

	return requests, nil
}

type request struct {
	days     int
	user     int
	pto_type int
}

func (s *PTOService) ApproveRequest(ctx *gin.Context, id int) error {
	var request request
	// get reqeust
	if err := s.db.QueryRow(ctx, "SELECT days, user_id, pto_type_id FROM pto_requests WHERE id = $1", id).Scan(&request.days, &request.user, &request.pto_type); err != nil {
		fmt.Printf("Error request does not exist\nDatabase error: %v", err)
		return err
	}

	// reduce user balance
	if _, err := s.db.Exec(ctx, "UPDATE pto_balances SET balance = balance - $1 WHERE user_id = $2 AND pto_type_id = $3", request.days, request.user, request.pto_type); err != nil {
		fmt.Printf("Error cant update PTO balance\nDatabase error: %v", err)
		return err
	}

	// set status to approved
	_, err := s.db.Exec(ctx, "UPDATE pto_requests SET status = $1 WHERE id = $2", StatusApproved, id)
	if err != nil {
		fmt.Printf("Error approving request\nDatabase error: %v", err)
		return err
	}

	return nil
}
