// Auto generated code. DO NOT EDIT.
package example

type UnitRepository interface {
	Create(unit *Unit) (int, error)
	Update(unit *Unit) error

	FindById(id int) (*Unit, error)
	FindRawById(id int) (*Unit, error)
	FindByUnitType(unitType string) ([]*Unit, error)

	MarkDeleted(id int) error
	FindByNameAndColor(name string, color string) ([]*Unit, error)
	// Nested objects
	LoadDetails(unit *Unit) error

	FindByDetailsDetailType(detailType string) ([]*Unit, error)
}

type UnitRepositoryCreator func(Transaction) (UnitRepository, error)
