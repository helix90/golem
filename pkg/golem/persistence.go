package golem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PersistentLearningManager manages persistent learning storage
type PersistentLearningManager struct {
	StoragePath  string
	BackupPath   string
	MaxBackups   int
	AutoSave     bool
	SaveInterval time.Duration
	lastSave     time.Time
}

// NewPersistentLearningManager creates a new persistent learning manager
func NewPersistentLearningManager(storagePath string) *PersistentLearningManager {
	return &PersistentLearningManager{
		StoragePath:  storagePath,
		BackupPath:   filepath.Join(storagePath, "backups"),
		MaxBackups:   5,
		AutoSave:     true,
		SaveInterval: 30 * time.Second,
		lastSave:     time.Now(),
	}
}

// PersistentCategory represents a category stored persistently
type PersistentCategory struct {
	Category  Category  `json:"category"`
	LearnedAt time.Time `json:"learned_at"`
	Source    string    `json:"source"`   // Source of the learning (user, system, etc.)
	Version   string    `json:"version"`  // Version of the learning system
	Checksum  string    `json:"checksum"` // Checksum for integrity verification
}

// PersistentLearningData represents the complete persistent learning data
type PersistentLearningData struct {
	Categories   []PersistentCategory `json:"categories"`
	LastUpdated  time.Time            `json:"last_updated"`
	Version      string               `json:"version"`
	TotalLearned int                  `json:"total_learned"`
}

// SavePersistentCategories saves categories to persistent storage
func (plm *PersistentLearningManager) SavePersistentCategories(categories []Category, source string) error {
	if plm.StoragePath == "" {
		return fmt.Errorf("storage path not configured")
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(plm.StoragePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Convert categories to persistent format
	persistentCategories := make([]PersistentCategory, len(categories))
	now := time.Now()

	for i, category := range categories {
		persistentCategories[i] = PersistentCategory{
			Category:  category,
			LearnedAt: now,
			Source:    source,
			Version:   "1.2.3", // Current version
			Checksum:  plm.calculateChecksum(category),
		}
	}

	// Create persistent learning data
	data := PersistentLearningData{
		Categories:   persistentCategories,
		LastUpdated:  now,
		Version:      "1.2.3",
		TotalLearned: len(persistentCategories),
	}

	// Save to file
	filename := filepath.Join(plm.StoragePath, "learned_categories.json")
	if err := plm.saveToFile(filename, data); err != nil {
		return fmt.Errorf("failed to save persistent categories: %v", err)
	}

	// Create backup
	if err := plm.createBackup(data); err != nil {
		// Log warning but don't fail the operation
		fmt.Printf("Warning: Failed to create backup: %v\n", err)
	}

	plm.lastSave = now
	return nil
}

// LoadPersistentCategories loads categories from persistent storage
func (plm *PersistentLearningManager) LoadPersistentCategories() ([]Category, error) {
	if plm.StoragePath == "" {
		return nil, fmt.Errorf("storage path not configured")
	}

	filename := filepath.Join(plm.StoragePath, "learned_categories.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []Category{}, nil // No persistent categories yet
	}

	// Load from file
	data, err := plm.loadFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load persistent categories: %v", err)
	}

	// Convert back to categories
	categories := make([]Category, len(data.Categories))
	for i, pc := range data.Categories {
		categories[i] = pc.Category
	}

	return categories, nil
}

// AppendPersistentCategory adds a single category to persistent storage
func (plm *PersistentLearningManager) AppendPersistentCategory(category Category, source string) error {
	// Load existing categories
	existingCategories, err := plm.LoadPersistentCategories()
	if err != nil {
		return fmt.Errorf("failed to load existing categories: %v", err)
	}

	// Check if category already exists (by pattern)
	normalizedPattern := NormalizePattern(category.Pattern)
	for i, existing := range existingCategories {
		if NormalizePattern(existing.Pattern) == normalizedPattern {
			// Update existing category
			existingCategories[i] = category
			return plm.SavePersistentCategories(existingCategories, source)
		}
	}

	// Add new category
	existingCategories = append(existingCategories, category)
	return plm.SavePersistentCategories(existingCategories, source)
}

// RemovePersistentCategory removes a category from persistent storage
func (plm *PersistentLearningManager) RemovePersistentCategory(category Category) error {
	// Load existing categories
	existingCategories, err := plm.LoadPersistentCategories()
	if err != nil {
		return fmt.Errorf("failed to load existing categories: %v", err)
	}

	// Find and remove the category
	normalizedPattern := NormalizePattern(category.Pattern)
	for i, existing := range existingCategories {
		if NormalizePattern(existing.Pattern) == normalizedPattern {
			// Remove the category
			existingCategories = append(existingCategories[:i], existingCategories[i+1:]...)
			return plm.SavePersistentCategories(existingCategories, "system")
		}
	}

	return fmt.Errorf("category not found: %s", normalizedPattern)
}

// GetPersistentCategoryInfo returns information about persistent categories
func (plm *PersistentLearningManager) GetPersistentCategoryInfo() (map[string]interface{}, error) {
	if plm.StoragePath == "" {
		return nil, fmt.Errorf("storage path not configured")
	}

	filename := filepath.Join(plm.StoragePath, "learned_categories.json")

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return map[string]interface{}{
			"total_categories": 0,
			"last_updated":     nil,
			"version":          "1.2.3",
		}, nil
	}

	// Load data
	data, err := plm.loadFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load persistent data: %v", err)
	}

	return map[string]interface{}{
		"total_categories": len(data.Categories),
		"last_updated":     data.LastUpdated,
		"version":          data.Version,
		"storage_path":     plm.StoragePath,
	}, nil
}

// saveToFile saves data to a JSON file
func (plm *PersistentLearningManager) saveToFile(filename string, data PersistentLearningData) error {
	// Create temporary file first
	tempFile := filename + ".tmp"

	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to encode data: %v", err)
	}

	// Close file before renaming
	file.Close()

	// Rename temp file to final file
	if err := os.Rename(tempFile, filename); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to rename temporary file: %v", err)
	}

	return nil
}

// loadFromFile loads data from a JSON file
func (plm *PersistentLearningManager) loadFromFile(filename string) (PersistentLearningData, error) {
	var data PersistentLearningData

	file, err := os.Open(filename)
	if err != nil {
		return data, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode data: %v", err)
	}

	return data, nil
}

// createBackup creates a backup of the persistent data
func (plm *PersistentLearningManager) createBackup(data PersistentLearningData) error {
	if plm.BackupPath == "" {
		return nil // No backup path configured
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(plm.BackupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(plm.BackupPath, fmt.Sprintf("learned_categories_%s.json", timestamp))

	// Save backup
	if err := plm.saveToFile(backupFile, data); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Clean up old backups
	plm.cleanupOldBackups()

	return nil
}

// cleanupOldBackups removes old backup files
func (plm *PersistentLearningManager) cleanupOldBackups() {
	if plm.MaxBackups <= 0 {
		return
	}

	// List backup files
	files, err := filepath.Glob(filepath.Join(plm.BackupPath, "learned_categories_*.json"))
	if err != nil {
		return // Ignore error
	}

	// If we have more than MaxBackups, remove the oldest ones
	if len(files) > plm.MaxBackups {
		// Sort by modification time (oldest first)
		// For simplicity, we'll just remove the first few files
		// In a production system, you'd want to sort by modification time
		for i := 0; i < len(files)-plm.MaxBackups; i++ {
			os.Remove(files[i])
		}
	}
}

// calculateChecksum calculates a simple checksum for a category
func (plm *PersistentLearningManager) calculateChecksum(category Category) string {
	// Simple checksum based on pattern and template
	content := category.Pattern + "|" + category.Template
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// ShouldAutoSave checks if it's time to auto-save
func (plm *PersistentLearningManager) ShouldAutoSave() bool {
	if !plm.AutoSave {
		return false
	}
	return time.Since(plm.lastSave) > plm.SaveInterval
}

// SetStoragePath sets the storage path
func (plm *PersistentLearningManager) SetStoragePath(path string) {
	plm.StoragePath = path
	plm.BackupPath = filepath.Join(path, "backups")
}

// SetAutoSave configures auto-save settings
func (plm *PersistentLearningManager) SetAutoSave(enabled bool, interval time.Duration) {
	plm.AutoSave = enabled
	plm.SaveInterval = interval
}
