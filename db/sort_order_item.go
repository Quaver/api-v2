package db

import (
	"fmt"
)

type Sortable interface {
	*MusicArtist | *PinnedScore
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
func CustomizeSortOrder[T Sortable](items []T, ids []int, updateOrder func(item T, sortOrder int) error) error {
	for i, id := range ids {
		for _, item := range items {
			switch v := any(item).(type) {
			case *PinnedScore:
				if id != v.ScoreId {
					continue
				}
			case *MusicArtist:
				if id != v.Id {
					continue
				}
			default:
				return fmt.Errorf("cannot customize sort order for type: %v", v)
			}

			if err := updateOrder(item, i); err != nil {
				return err
			}
		}
	}

	return nil
}
