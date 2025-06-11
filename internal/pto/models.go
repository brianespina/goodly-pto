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

type PTOAction string

const (
	ActionApprove PTOAction = "approve"
	ActionCancel  PTOAction = "cancel"
	ActionDeny    PTOAction = "deny"
)

type PTOType string

const (
	TypeVacation PTOType = "vacation leave"
	TypeSick     PTOType = "sick leave"
	TypeAll      PTOType = "all"
)

type PTODate string

const (
	DateAll       PTODate = "all"
	DateUpcomming PTODate = "upcomming"
	DatePast      PTODate = "past"
)

type PTOListConfig struct {
	Hide   string
	Action []PTOAction
}

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
