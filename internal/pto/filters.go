package pto

type PTOFilters struct {
	Status PTOStatus
	Type   PTOType
	Date   PTODate
	View   PTOListView
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

func WithView(view PTOListView) PTOOption {
	return func(f *PTOFilters) {
		f.View = view
	}
}

func ApplyFilters(opts []PTOOption) PTOFilters {
	filters := PTOFilters{
		Status: StatusPending,
		Type:   TypeAll,
		Date:   DateUpcomming,
		View:   ListMyView,
	}

	for _, opt := range opts {
		opt(&filters)
	}
	return filters
}
