package model

type User struct {
	Id       int
	Uuid     string
	Name     string
	Password string
	Question string
	Answer   string
}

type ChangePassword struct {
	Name        string
	OldPassword string
	NewPassword string
}
