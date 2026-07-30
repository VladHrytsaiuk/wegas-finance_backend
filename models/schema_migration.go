package models

// SchemaMigration records one-time data migrations. It is separate from GORM's
// AutoMigrate so data transformations are never repeated on later starts.
type SchemaMigration struct {
	ID        string `gorm:"primaryKey"`
	AppliedAt int64  `json:"applied_at"`
}
