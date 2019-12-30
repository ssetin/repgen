// Auto generated code. DO NOT EDIT.
package mocks

import (
	"errors"
	"sync"
	"time"

	"github.com/ssetin/repgen/example"
)

type keyUnit struct {
	mux     sync.Mutex
	counter int
}

func (k *keyUnit) nextKey() int {
	k.mux.Lock()
	defer k.mux.Unlock()
	k.counter++
	return k.counter
}

// UnitRepositoryMock =======================================================
type UnitRepositoryMock struct {
	mux  sync.Mutex
	data map[int]*example.Unit

	details map[int]*example.Detail
}

var unitRepository *UnitRepositoryMock
var keysGeneratorUnit keyUnit

func NewUnitRepository(_ example.Transaction) (example.UnitRepository, error) {
	if unitRepository == nil {
		unitRepository = &UnitRepositoryMock{
			data:    make(map[int]*example.Unit),
			details: make(map[int]*example.Detail),
		}
	}
	return unitRepository, nil
}

func ClearUnitRepositoryMock() {
	unitRepository = nil
}

func (ar UnitRepositoryMock) copy(unit *example.Unit) *example.Unit {
	return &example.Unit{
		Id:       unit.Id,
		UnitType: unit.UnitType,
		Name:     unit.Name,
		Color:    unit.Color,
		Deleted:  unit.Deleted,
	}
}

func (ar UnitRepositoryMock) updateNestedObjects(unit *example.Unit) error {
	var err error
	for _, item := range unit.Details {
		_, err = ar.createOrUpdateDetail(item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ar *UnitRepositoryMock) Create(unit *example.Unit) (int, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()

	newItem := ar.copy(unit)

	if unit.Id == nil {
		newKey := keysGeneratorUnit.nextKey()
		newItem.Id = &newKey
	}

	if item, ok := ar.data[*newItem.Id]; ok {
		return *item.Id, errors.New("record already exists")
	}

	ar.data[*newItem.Id] = newItem

	err := ar.updateNestedObjects(unit)
	if err != nil {
		return *newItem.Id, err
	}

	return *newItem.Id, nil
}

func (ar *UnitRepositoryMock) Update(unit *example.Unit) error {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	if _, ok := ar.data[*unit.Id]; !ok {
		return errors.New("record not found")
	}
	ar.data[*unit.Id] = unit

	err := ar.updateNestedObjects(unit)
	if err != nil {
		return err
	}

	return nil
}

func (ar *UnitRepositoryMock) FindRawById(id int) (*example.Unit, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	if item, ok := ar.data[id]; ok {
		return ar.copy(item), nil
	}
	return nil, nil
}

func (ar *UnitRepositoryMock) FindById(id int) (*example.Unit, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	if item, ok := ar.data[id]; ok && item.IsActive() {
		return ar.copy(item), nil
	}
	return nil, nil
}

func (ar *UnitRepositoryMock) FindByUnitType(unitType string) ([]*example.Unit, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	var filtered []*example.Unit

	for _, item := range ar.data {
		if item.UnitType == unitType && item.IsActive() {
			filtered = append(filtered, ar.copy(item))
		}
	}
	return filtered, nil
}

func (ar *UnitRepositoryMock) MarkDeleted(id int) error {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	if item, ok := ar.data[id]; ok {
		now := time.Now()
		item.Deleted = &now
	}
	return nil
}

func (ar *UnitRepositoryMock) FindByNameAndColor(name string, color string) ([]*example.Unit, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	var filtered []*example.Unit

	for _, item := range ar.data {
		if item.Name == name &&
			item.Color == color && item.IsActive() {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// Nested Details =======================================================
func (ar UnitRepositoryMock) createOrUpdateDetail(detail *example.Detail) (int, error) {
	newItem := ar.copyDetail(detail)

	if detail.Id == nil {
		newKey := keysGeneratorUnit.nextKey()
		newItem.Id = &newKey
	}

	ar.details[*newItem.Id] = newItem
	return *newItem.Id, nil
}

func (ar UnitRepositoryMock) copyDetail(detail *example.Detail) *example.Detail {
	return &example.Detail{
		Id:          detail.Id,
		DetailType:  detail.DetailType,
		Description: detail.Description,
	}
}

func (ar UnitRepositoryMock) LoadDetails(unit *example.Unit) error {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	for _, item := range ar.details {
		if *item.Id == *unit.Id {
			unit.Details = append(unit.Details, item)
		}
	}
	return nil
}

func (ar UnitRepositoryMock) FindByDetailsDetailType(detailType string) ([]*example.Unit, error) {
	ar.mux.Lock()
	defer ar.mux.Unlock()
	var filtered []*example.Unit

	for _, item := range ar.details {
		if item.DetailType == detailType {
			if foundItem, ok := ar.data[*item.Id]; ok {
				filtered = append(filtered, ar.copy(foundItem))
			}
		}
	}
	return filtered, nil
}
