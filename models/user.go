package models

type User struct {
	Name  string
	Email string
	Role  string
}

type Role struct {
	Title string
}
