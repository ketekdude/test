package main

import (
	"fmt"
	company "test/oop/companyadapter"
	"test/oop/employee"
)

func main() {
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
