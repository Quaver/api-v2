package db

// SyncSortOrder Generic function to update the sort order of a slice of items
func SyncSortOrder[T any](items []T, updateOrder func(item T, sortOrder int) error) error {
	for i, item := range items {
		if err := updateOrder(item, i); err != nil {
			return err
		}
	}

	return nil
}
