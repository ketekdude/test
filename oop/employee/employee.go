package employee

import (
	"fmt"

	"github.com/google/uuid"
)

// Inherit company func in this package
type Employee struct {
	Id      string
	Name    string
	Company companyInterface
}

type companyInterface interface {
	ChangeRegion(string) string
}

func NewEmployee(name string, country string, company companyInterface) (*Employee, error) {
	if name == "" || country == "" {
		return nil, fmt.Errorf("name empty")
	}

	if company == nil {
		return nil, fmt.Errorf("ERROR BANGKE")
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
