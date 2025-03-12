package animal

// Define the basic animal interface
type Animal interface {
	MakeSound()
	Eat()
}

// Define the flying animal interface that extends Animal
type FlyingAnimal interface {
	Animal
	Fly()
}

// Factory function
func NewAnimal(animalType string) Animal {
	switch animalType {
	case "chicken":
		return &Chicken{}
	case "dog":
		return &Dog{}
	case "eagle":
		return &Eagle{}
	default:
		panic("Unknown animal")
	}
}
