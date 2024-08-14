package db

type Sortable interface {
	*MusicArtist | *PinnedScore
	ID() int
}

// SyncSortOrder Generic function to update the sort order of a slice of items
func SyncSortOrder[T any](items []T, updateOrder func(item T, sortOrder int) error) error {
	for i, item := range items {
		if err := updateOrder(item, i); err != nil {
			return err
		}
	}

	return nil
}

// CustomizeSortOrder Customizes the sort order of values
func CustomizeSortOrder[T Sortable](items []T, ids []int, updateOrder func(item T, sortOrder int) error,
	syncSort func() error) error {
	for i, id := range ids {
		for _, item := range items {
			if id != item.ID() {
				continue
			}

			if err := updateOrder(item, i); err != nil {
				return err
			}
		}
	}

	return syncSort()
}
