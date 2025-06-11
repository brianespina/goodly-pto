package pto

type PTOFilters struct {
	Status PTOStatus
	Type   PTOType
	Date   PTODate
}

type PTOOption func(*PTOFilters)

func WithStatus(status PTOStatus) PTOOption {
	return func(f *PTOFilters) {
		f.Status = status
	}
}

func WithType(ptoType PTOType) PTOOption {
	return func(f *PTOFilters) {
		f.Type = ptoType
	}
}

func WithDate(date PTODate) PTOOption {
	return func(f *PTOFilters) {
		f.Date = date
	}
}
