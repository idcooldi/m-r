package SimpleCQRS

import (
	"time"
)

type InventoryCommandHandlers struct {
	repo InventoryItemRepository
}

func NewInventoryCommandHandlers(repo InventoryItemRepository) InventoryCommandHandlers {
	return InventoryCommandHandlers{repo}
}

func (r *InventoryCommandHandlers) HandleCreateInventoryItem(m Command) error {
	message := m.(CreateInventoryItem)
	item := NewInventoryItem(message.InventoryItemId, message.Name)
	return r.repo.Save(item, -1)
}

func (r *InventoryCommandHandlers) HandleDeactivateInventoryItem(m Command) error {
	message := m.(DeactivateInventoryItem)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	err := item.Deactivate()
	if err != nil {
		return err
	}
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleRemoveItemsFromInventory(m Command) error {
	message := m.(RemoveItemsFromInventory)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	err := item.Remove(message.Count)
	if err != nil {
		return err
	}
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleCheckInItemsToInventory(m Command) error {
	message := m.(CheckInItemsToInventory)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	err := item.CheckIn(message.Count)
	if err != nil {
		return err
	}
	return r.repo.Save(item, message.OriginalVersion)
}

func (r *InventoryCommandHandlers) HandleRenameInventoryItem(m Command) error {
	message := m.(RenameInventoryItem)
	ar, _ := r.repo.GetById(message.InventoryItemId)
	item := ar.(*InventoryItem)
	//time.Sleep(10 * time.Second) // Name changes take ages, not sure why
	time.Sleep(2 * time.Millisecond) // This is short enough to be handled before my next HTTP request (i.e. it "looks" synchronous)
	err := item.ChangeName(message.NewName)
	if err != nil {
		return err
	}
	return r.repo.Save(item, message.OriginalVersion)
}
