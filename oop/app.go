package main

import (
	"fmt"
	"test/oop/animal"
	company "test/oop/companyadapter"
	"test/oop/employee"
)

func main() {
	a := animal.NewAnimal("chicken")
	b := animal.NewAnimal("dog")
	c := animal.NewAnimal("eagle")

	if flyingAnimal, ok := c.(animal.FlyingAnimal); ok {
		flyingAnimal.Fly() // Call Fly() only for FlyingAnimal types
	} else {
		fmt.Println("This animal cannot fly.")
	}

	test, ok := b.(*animal.Dog)
	if ok {
		test.Test()
	}
	// } else {
	// 	test.Test()
	// }

	a.MakeSound()
	// a.Test()
	b.MakeSound()
}

func testOOP1() {
	c, err := company.NewCompany("tokopedia", "indonesia")
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}
	c.ChangeRegion("asia tenggara")
	fmt.Println(c.GetCompanyName(), c.GetCompanyCountry())

	e, err := employee.NewEmployee("rickigozal", "indonesia", c)
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}
	e.Company.ChangeRegion("asia timur")
	fmt.Println(e.GetEmployeeName())
	fmt.Println(c.GetRegion())
	fmt.Println(checkInterfaceName(c))
}

// abstraction func
func checkInterfaceName(c company.CompanyInterface) string {
	return c.GetRegion()
}
