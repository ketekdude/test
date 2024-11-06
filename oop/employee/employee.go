package employee

import (
	"fmt"
	"test/oop/company"

	"github.com/google/uuid"
)

// Inherit company func in this package
type Employee struct {
	Id      string
	Name    string
	Company *company.Company
}

func NewEmployee(name string, country string, company *company.Company) (*Employee, error) {
	if name == "" || country == "" {
		return nil, fmt.Errorf("name empty")
	}

	return &Employee{
		Id:      uuid.New().String(),
		Name:    name,
		Company: company,
	}, nil
}

func (c *Employee) GetEmployeeName() string {
	return c.Name
}