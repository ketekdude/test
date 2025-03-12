package animal

import "fmt"

// Concrete implementations
type Eagle struct {
	FlyingAnimal
}

func (c *Eagle) MakeSound() {
	fmt.Println("kukuruyuk")
}

func (c *Eagle) Eat() {
	fmt.Println("beras")
}

func (c *Eagle) Test() {
	fmt.Println("beras chicken")
}

func (c *Eagle) Fly() {
	fmt.Println("terbang cuy!")
}
