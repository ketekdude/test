package animal

import "fmt"

// Concrete implementations
type Dog struct {
	Chicken
}

func (c Dog) MakeSound() {
	fmt.Println("ANJENK")
}

func (c Dog) Eat() {
	fmt.Println("TULANG")
}
