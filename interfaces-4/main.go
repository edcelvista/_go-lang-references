package main

import "fmt"

type Service struct{}
type Users struct{}

func (s *Service) Users() UserInterface {
	return &Users{}
}

func (u *Users) Watch() string {
	return "Watching users..."
}

type ServiceInterface interface {
	Users() UserInterface
}

type UserInterface interface {
	Watch() string
}

func main() {
	s := &Service{}
	result := s.Users().Watch()
	fmt.Println(result)
}
