package main

import "fmt"

func main() {
	user := newUser("Alice")
	fmt.Println(user.name)
}

func newUser(name string) *User {
	user := &User{name: name}
	return user
}

type User struct {
	name string
}
