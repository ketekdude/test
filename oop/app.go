package main

import (
	"fmt"
	"test/oop/company"
	"test/oop/employee"
)

func main() {
	c, err := company.NewCompany("tokopedia", "indonesia")
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}

	c.Name = "tokopedia | shop"

	fmt.Println(c.GetCompanyName(), c.GetCompanyCountry())

	e, err := employee.NewEmployee("rickigozal", "indonesia", c)
	if err != nil {
		fmt.Println("ERROR WOI", err)
		return
	}

	fmt.Println(e.GetEmployeeName(), e.Company.GetCompanyName())
}
