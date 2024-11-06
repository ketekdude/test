package companyadapter

import (
	"fmt"
	"test/oop/company"
)

type CompanyInterface interface {
	GetCompanyName() string
	ChangeRegion(string) string
	GetCompanyCountry() string
	GetRegion() string
}

func NewCompany(name string, country string) (CompanyInterface, error) {
	if name == "" || country == "" {
		return nil, fmt.Errorf("name empty")
	}
	var c company.ICompany

	c, err := company.NewCompany(name, country)
	if err != nil {
		fmt.Println("fail init company adapter")
	}
	return c, nil
}
