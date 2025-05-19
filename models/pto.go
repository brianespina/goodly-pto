package models

type PTOStatus string

const (
	StatusPending  PTOStatus = "pending"
	StatusApproved PTOStatus = "approved"
	StatusDenied   PTOStatus = "denied"
)

type PTORequest struct {
	Id        int
	User      int
	Type      int
	StartDate string
	EndDate   string
	Hours     float64
	Days      float64
	Status    PTOStatus
}
