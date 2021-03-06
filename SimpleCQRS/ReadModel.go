package SimpleCQRS

import (
	"errors"
	"sync"
)

type InventoryItemDetailsDto struct {
	Id           Guid
	Name         string
	CurrentCount int
	Version      int
}

type InventoryItemListDto struct {
	Id   Guid
	Name string
}

type BSDB struct {
	list    []InventoryItemListDto
	details map[Guid]InventoryItemDetailsDto
	s       sync.RWMutex
}

func NewBSDB() BSDB {
	return BSDB{list: make([]InventoryItemListDto, 0), details: make(map[Guid]InventoryItemDetailsDto)}
}

type ReadModel interface {
	GetInventoryItems() []InventoryItemListDto
	GetInventoryItemDetails(id Guid) (InventoryItemDetailsDto, error)
}

type InventoryItemDetailView struct {
	db *BSDB
}

func NewInventoryItemDetailView(bsdb *BSDB) InventoryItemDetailView {
	return InventoryItemDetailView{bsdb}
}

func (detail *InventoryItemDetailView) ProcessInventoryItemCreated(e Event) error {
	evt, ok := e.(InventoryItemCreated)
	if !ok {
		return errors.New("passed incorrect event")
	}
	detail.db.s.Lock()
	defer detail.db.s.Unlock()

	detail.db.details[evt.Id()] = InventoryItemDetailsDto{evt.Id(), evt.Name(), 0, 0}
	return nil
}
func (detail *InventoryItemDetailView) ProcessInventoryItemDeactivated(e Event) error {
	evt, ok := e.(InventoryItemDeactivated)
	if !ok {
		return errors.New("passed incorrect event")
	}
	detail.db.s.Lock()
	defer detail.db.s.Unlock()

	delete(detail.db.details, evt.Id())
	return nil
}
func (detail *InventoryItemDetailView) ProcessInventoryItemRenamed(e Event) error {
	evt, ok := e.(InventoryItemRenamed)
	if !ok {
		return errors.New("passed incorrect event")
	}
	detail.db.s.Lock()
	defer detail.db.s.Unlock()

	original, ok := detail.db.details[evt.Id()]
	if !ok {
		return errors.New("this should never happen")
	}
	original.Name = evt.NewName()
	original.Version = evt.Version()
	detail.db.details[evt.Id()] = original
	return nil
}
func (detail *InventoryItemDetailView) ProcessItemsCheckedInToInventory(e Event) error {
	evt, ok := e.(ItemsCheckedInToInventory)
	if !ok {
		return errors.New("passed incorrect event")
	}
	detail.db.s.Lock()
	defer detail.db.s.Unlock()

	original, ok := detail.db.details[evt.Id()]
	if !ok {
		return errors.New("this should never happen")
	}
	original.CurrentCount += evt.Count()
	original.Version = evt.Version()
	detail.db.details[evt.Id()] = original
	return nil
}
func (detail *InventoryItemDetailView) ProcessItemsRemovedFromInventory(e Event) error {
	evt, ok := e.(ItemsRemovedFromInventory)
	if !ok {
		return errors.New("passed incorrect event")
	}
	detail.db.s.Lock()
	defer detail.db.s.Unlock()

	original, ok := detail.db.details[evt.Id()]
	if !ok {
		return errors.New("this should never happen")
	}
	original.CurrentCount -= evt.Count()
	original.Version = evt.Version()
	detail.db.details[evt.Id()] = original
	return nil
}

type InventoryItemListView struct {
	db *BSDB
}

func NewInventoryListView(bsdb *BSDB) InventoryItemListView {
	return InventoryItemListView{bsdb}
}
func (list *InventoryItemListView) ProcessInventoryItemCreated(e Event) error {
	evt, ok := e.(InventoryItemCreated)
	if !ok {
		return errors.New("passed incorrect event")
	}
	list.db.s.Lock()
	defer list.db.s.Unlock()

	list.db.list = append(list.db.list, InventoryItemListDto{evt.Id(), evt.Name()})
	return nil
}

func (list *InventoryItemListView) ProcessInventoryItemRenamed(e Event) error {
	evt, ok := e.(InventoryItemRenamed)
	if !ok {
		return errors.New("passed incorrect event")
	}
	list.db.s.Lock()
	defer list.db.s.Unlock()

	for i, item := range list.db.list {
		if evt.Id() == item.Id {
			list.db.list[i] = InventoryItemListDto{evt.Id(), evt.NewName()}
		}
	}
	return nil
}

func (list *InventoryItemListView) ProcessInventoryItemDeactivated(e Event) error {
	evt, ok := e.(InventoryItemDeactivated)
	if !ok {
		return errors.New("passed incorrect event")
	}
	list.db.s.Lock()
	defer list.db.s.Unlock()

	newList := make([]InventoryItemListDto, 0)
	for _, item := range list.db.list {
		if evt.Id() != item.Id {
			newList = append(newList, item)
		}
	}
	list.db.list = newList
	return nil
}

type ReadModelFacade struct {
	db *BSDB
}

func NewReadModelFacade(db *BSDB) ReadModelFacade {
	return ReadModelFacade{db}
}

func (rmf *ReadModelFacade) GetInventoryItems() []InventoryItemListDto {
	rmf.db.s.RLock()
	defer rmf.db.s.RUnlock()

	tmp := make([]InventoryItemListDto, len(rmf.db.list))
	copy(tmp, rmf.db.list)
	return tmp
}

func (rmf *ReadModelFacade) GetInventoryItemDetails(id Guid) (InventoryItemDetailsDto, error) {
	rmf.db.s.RLock()
	defer rmf.db.s.RUnlock()

	item, ok := rmf.db.details[id]
	if !ok {
		return item, errors.New("No item")
	}
	return item, nil
}
