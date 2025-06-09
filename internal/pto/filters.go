package pto

type PTOFilters struct {
	Status PTOStatus
	Type   PTOType
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
