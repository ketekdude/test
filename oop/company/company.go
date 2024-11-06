package company

import (
	"fmt"

	"github.com/google/uuid"
)

// Encapsulate the func to this struct on OOP practice
type Company struct {
	Id      string
	Name    string
	Country string
	Region  string
}

func NewCompany(name string, country string) (*Company, error) {
	if name == "" || country == "" {
		return nil, fmt.Errorf("name empty")
	}

	return &Company{
		Id:      uuid.New().String(),
		Name:    name,
		Country: country,
		Region:  "asia",
	}, nil
}

func (c *Company) GetCompanyName() string {
	return c.Name
}

func (c *Company) GetCompanyCountry() string {
	return c.Country
}

func (c *Company) GetRegion() string {
	return c.Region
}
