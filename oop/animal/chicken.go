package animal

import "fmt"

// Concrete implementations
type Chicken struct {
	Animal
}

func (c *Chicken) MakeSound() {
	fmt.Println("kukuruyuk")
}

func (c *Chicken) Eat() {
	fmt.Println("beras")
}

func (c *Chicken) Test() {
	fmt.Println("beras chicken TEST")
}
