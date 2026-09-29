package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Role string

type ProductStatus string

type OrderStatus string

const (
	RoleCustomer Role = "CUSTOMER"
	RoleAdmin    Role = "ADMIN"

	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusInactive ProductStatus = "INACTIVE"
	ProductStatusArchived ProductStatus = "ARCHIVED"

	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusConfirmed  OrderStatus = "CONFIRMED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         Role      `gorm:"type:varchar(20);not null;default:CUSTOMER" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name              string          `gorm:"not null" json:"name"`
	SKU               string          `gorm:"uniqueIndex;not null" json:"sku"`
	Description       string          `json:"description"`
	CategoryID        uuid.UUID       `gorm:"type:uuid;not null" json:"category_id"`
	Category          Category        `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Price             decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"price"`
	StockQuantity     int32           `gorm:"not null;default:0" json:"stock_quantity"`
	LowStockThreshold int32           `gorm:"not null;default:5" json:"low_stock_threshold"`
	Status            ProductStatus   `gorm:"type:varchar(20);default:ACTIVE" json:"status"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Images            []ProductImage  `gorm:"foreignKey:ProductID" json:"images,omitempty"`
}

type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	ImageURL  string    `gorm:"not null" json:"image_url"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type Cart struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Items     []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null" json:"cart_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

type Order struct {
	ID              uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          uuid.UUID    `gorm:"type:uuid;not null" json:"user_id"`
	OrderNumber     string       `gorm:"uniqueIndex;not null" json:"order_number"`
	Status          OrderStatus  `gorm:"type:varchar(20);default:PENDING" json:"status"`
	Subtotal        decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"subtotal"`
	Tax             decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"tax"`
	ShippingFee     decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"shipping_fee"`
	Total           decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"total"`
	ShippingAddress string       `gorm:"type:text;not null" json:"shipping_address"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Items           []OrderItem  `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID     uuid.UUID       `gorm:"type:uuid;not null" json:"order_id"`
	ProductID   uuid.UUID       `gorm:"type:uuid;not null" json:"product_id"`
	ProductName string          `gorm:"not null" json:"product_name"`
	SKU         string          `gorm:"not null" json:"sku"`
	UnitPrice   decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Quantity    int             `gorm:"not null" json:"quantity"`
	LineTotal   decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"line_total"`
	CreatedAt   time.Time       `json:"created_at"`
}

type InventoryMovement struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	ChangeQuantity int32     `gorm:"not null" json:"change_quantity"`
	PreviousQty    int32     `gorm:"not null" json:"previous_quantity"`
	NewQty         int32     `gorm:"not null" json:"new_quantity"`
	Reason         string    `gorm:"not null" json:"reason"`
	AdminUserID    *uuid.UUID `gorm:"type:uuid" json:"admin_user_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ActorUserID *uuid.UUID `gorm:"type:uuid" json:"actor_user_id,omitempty"`
	Action      string    `gorm:"not null" json:"action"`
	EntityType  string    `gorm:"not null" json:"entity_type"`
	EntityID    string    `gorm:"not null" json:"entity_id"`
	Metadata    string    `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateProductRequest struct {
	Name              string    `json:"name" binding:"required,min=2,max=200"`
	SKU               string    `json:"sku" binding:"required,max=64"`
	Description       string    `json:"description" binding:"required,min=10"`
	CategoryID        string    `json:"category_id" binding:"required,uuid"`
	Price             string    `json:"price" binding:"required"`
	StockQuantity     int32     `json:"stock_quantity" binding:"min=0"`
	LowStockThreshold int32     `json:"low_stock_threshold" binding:"min=0"`
	Status            string    `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE ARCHIVED"`
	ImageURLs         []string  `json:"image_urls"`
}

type UpdateProductRequest struct {
	Name              string `json:"name"`
	SKU               string `json:"sku"`
	Description       string `json:"description"`
	CategoryID        string `json:"category_id"`
	Price             string `json:"price"`
	StockQuantity     int32  `json:"stock_quantity"`
	LowStockThreshold int32  `json:"low_stock_threshold"`
	Status            string `json:"status"`
	ImageURLs         []string `json:"image_urls"`
}

type CartAddRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required,min=10"`
	IdempotencyKey  string `json:"idempotency_key" binding:"required,min=8"`
}

type AuthResponse struct {
	User          User   `json:"user"`
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
