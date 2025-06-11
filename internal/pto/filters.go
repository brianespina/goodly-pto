package pto

type PTOFilters struct {
	Status PTOStatus
	Type   PTOType
	Date   PTODate
}

type PTOOption func(*PTOFilters)

func WithStatus(status PTOStatus) PTOOption {
	if status == "" {
		status = StatusPending
	}
	return func(f *PTOFilters) {
		f.Status = status
	}
}

func WithType(ptoType PTOType) PTOOption {
	if ptoType == "" {
		ptoType = TypeAll
	}
	return func(f *PTOFilters) {
		f.Type = ptoType
	}
}

func WithDate(date PTODate) PTOOption {
	if date == "" {
		date = DateUpcomming
	}
	return func(f *PTOFilters) {
		f.Date = date
	}
}
