package database

import (
	"fmt"
	"log"
	"time"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	log.Printf("Connecting to database: %s", cfg.DatabaseURL)

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductImage{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.InventoryMovement{},
		&models.AuditLog{},
		&models.RefreshToken{},
	); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	if err := seedCategories(db); err != nil {
		return nil, fmt.Errorf("seed categories failed: %w", err)
	}
	if err := seedProducts(db); err != nil {
		return nil, fmt.Errorf("seed products failed: %w", err)
	}

	// Force connection ping
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db failed: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	return db, nil
}

func seedCategories(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Category{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []models.Category{
		{Name: "Nutraceuticals", Description: "Performance and wellness nutrition"},
		{Name: "Industrial Chemicals", Description: "Process and technical chemical inputs"},
		{Name: "Pharmaceutical Ingredients", Description: "API and pharma excipient supply"},
		{Name: "Herbal Extracts", Description: "Plant-based therapeutic formulations"},
	}
	return db.Create(&categories).Error
}

func seedProducts(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	categories := []models.Category{}
	if err := db.Order("name").Find(&categories).Error; err != nil {
		return err
	}
	if len(categories) == 0 {
		return nil
	}

	products := []models.Product{
		{Name: "Vitamin D3 Plus", SKU: "NUT-VD3-001", Description: "High-potency vitamin D3 and K2 support for bone and immune health.", CategoryID: categories[0].ID, Price: decimalFromString("28.00"), StockQuantity: 120, LowStockThreshold: 12, Status: models.ProductStatusActive},
		{Name: "ImmunoBoost Capsules", SKU: "NUT-IMM-002", Description: "Phytochemical blend promoting immunity and daily vitality.", CategoryID: categories[0].ID, Price: decimalFromString("34.90"), StockQuantity: 85, LowStockThreshold: 10, Status: models.ProductStatusActive},
		{Name: "Process Stabilizer 40", SKU: "IND-PS40-101", Description: "Industrial stabilizer for water treatment and process flows.", CategoryID: categories[1].ID, Price: decimalFromString("145.50"), StockQuantity: 25, LowStockThreshold: 8, Status: models.ProductStatusActive},
		{Name: "Pharma Grade API 99", SKU: "PHAR-API-220", Description: "Analytical grade active ingredient for regulated formulations.", CategoryID: categories[2].ID, Price: decimalFromString("520.00"), StockQuantity: 18, LowStockThreshold: 6, Status: models.ProductStatusActive},
		{Name: "Ashwagandha Root Extract", SKU: "HERB-AW-315", Description: "Standardized adaptogenic root extract for stress resilience.", CategoryID: categories[3].ID, Price: decimalFromString("42.00"), StockQuantity: 70, LowStockThreshold: 15, Status: models.ProductStatusActive},
	}
	if err := db.Create(&products).Error; err != nil {
		return err
	}

	imageMap := map[string][]string{
		"NUT-VD3-001":  {"/images/supplements.jpg", "/images/Ayurvedic_Medicine.webp", "/images/Ayurvedic_herbs_roots_732x549_thumb-732x549.avif"},
		"NUT-IMM-002":  {"/images/Ayurvedic_Medicine.webp", "/images/brahmi.jpg", "/images/triphala.jpg"},
		"IND-PS40-101": {"/images/images.jpg", "/images/images (1).jpg", "/images/images (2).jpg"},
		"PHAR-API-220": {"/images/Ayurvedic-Raw-Materials.jpg", "/images/Raw-Drugs-1-scaled.jpg", "/images/DPP_171-scaled.jpg"},
		"HERB-AW-315":  {"/images/ashwagandha.jpg", "/images/mulethi.jpg", "/images/shilajit.jpg"},
	}

	for _, product := range products {
		imageURLs, ok := imageMap[product.SKU]
		if !ok {
			continue
		}
		for idx, imageURL := range imageURLs {
			if err := db.Create(&models.ProductImage{ProductID: product.ID, ImageURL: imageURL, SortOrder: idx}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func decimalFromString(s string) decimal.Decimal {
	v, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return v
}
