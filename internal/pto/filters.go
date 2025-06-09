package pto

type PTOFilters struct {
	Status PTOStatus
}

type PTOOption func(*PTOFilters)

func WithStatus(status PTOStatus) PTOOption {
	return func(f *PTOFilters) {
		f.Status = status
	}
}
