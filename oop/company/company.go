package company

import (
	"fmt"

	"github.com/google/uuid"
)

// Encapsulate the func to this struct on OOP practice
// do not let the parameter can be seen ouutside this package
// to alter / get the value, please use the given function
type Company struct {
	Id      string
	name    string
	country string
	region  string
}

func NewCompany(name string, country string) (*Company, error) {
	if name == "" || country == "" {
		return nil, fmt.Errorf("name empty")
	}

	return &Company{
		Id:      uuid.New().String(),
		name:    name,
		country: country,
		region:  "asia",
	}, nil
}

func (c *Company) GetCompanyName() string {
	return c.name
}

func (c *Company) GetCompanyCountry() string {
	return c.country
}

func (c *Company) GetRegion() string {
	return c.region
}

func (c *Company) ChangeRegion(region string) string {
	c.region = region
	return c.region
}
