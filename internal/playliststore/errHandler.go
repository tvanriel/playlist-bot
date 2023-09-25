package playliststore

import "gorm.io/gorm"

func tryQuery(tx *gorm.DB) error {
	return tx.Error
}
