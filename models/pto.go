package models

type PTOStatus string

const (
	StatusPending  PTOStatus = "pending"
	StatusApproved PTOStatus = "approved"
	StatusDenied   PTOStatus = "denied"
)

type PTORequest struct {
	Id        int
	User      string
	Type      string
	StartDate string
	EndDate   string
	Hours     float64
	Days      float64
	Reason    string
	Status    PTOStatus
}
