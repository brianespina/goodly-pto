package pto

import "time"

type PTOStatus string

const (
	StatusPending  PTOStatus = "pending"
	StatusApproved PTOStatus = "approved"
	StatusDenied   PTOStatus = "denied"
	StatusCanceled PTOStatus = "canceled"
	StatusAll      PTOStatus = "all"
)

type PTORequest struct {
	Id          int
	User        string
	Type        string
	StartDate   time.Time
	EndDate     time.Time
	CreatedDate time.Time
	Hours       float64
	Days        float64
	Reason      string
	Status      PTOStatus
}
