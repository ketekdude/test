package main

import (
	"fmt"
	"test/oop/company"
	"test/oop/employee"
)

type companyInterface interface {
	GetCompanyName() string
}

func main() {
	c, err := company.NewCompany("tokopedia", "indonesia")
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}
	c.ChangeRegion("asia tenggara")
	fmt.Println(c.GetCompanyName(), c.GetCompanyCountry(), c.GetRegion())

	e, err := employee.NewEmployee("rickigozal", "indonesia", c)
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}

	fmt.Println(e.GetEmployeeName(), e.Company.GetCompanyName(), e.Company.GetRegion())

	fmt.Println(checkInterfaceName(c))
}

// abstraction func
func checkInterfaceName(c companyInterface) string {
	return c.GetCompanyName()
}
