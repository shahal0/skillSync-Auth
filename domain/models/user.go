package model

type User struct {
	ID            string
	Email         string
	Password      string
	Role          string // candidate or employer
	Name          string
	Phone         string
	IsVerified    bool
}
