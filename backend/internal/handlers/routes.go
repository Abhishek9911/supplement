package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"suppliment/backend/internal/config"
	"suppliment/backend/internal/middleware"
	"suppliment/backend/internal/models"
	"suppliment/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type API struct {
	DB  *gorm.DB
	Cfg config.Config
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	api := &API{DB: db, Cfg: cfg}
	v1 := r.Group("/api/v1")

	v1.POST("/auth/register", api.Register)
	v1.POST("/auth/login", api.Login)
	v1.POST("/auth/refresh", api.Refresh)
	v1.POST("/auth/logout", middleware.AuthRequired(cfg), api.Logout)
	v1.GET("/products", api.ListProducts)
	v1.GET("/products/:id", api.GetProduct)
	v1.GET("/categories", api.ListCategories)
	v1.GET("/cart", middleware.AuthRequired(cfg), api.GetCart)
	v1.POST("/cart/items", middleware.AuthRequired(cfg), api.AddToCart)
	v1.PATCH("/cart/items/:id", middleware.AuthRequired(cfg), api.UpdateCartItem)
	v1.DELETE("/cart/items/:id", middleware.AuthRequired(cfg), api.RemoveCartItem)
	v1.POST("/orders/checkout", middleware.AuthRequired(cfg), api.Checkout)
	v1.GET("/orders", middleware.AuthRequired(cfg), api.ListOrders)
	v1.GET("/orders/:id", middleware.AuthRequired(cfg), api.GetOrder)
	v1.POST("/orders/:id/cancel", middleware.AuthRequired(cfg), api.CancelOrder)

	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(cfg), middleware.RequireRole(models.RoleAdmin))
	admin.GET("/overview", api.AdminOverview)
	admin.GET("/products", api.AdminListProducts)
	admin.POST("/products", api.AdminCreateProduct)
	admin.PUT("/products/:id", api.AdminUpdateProduct)
	admin.DELETE("/products/:id", api.AdminDeleteProduct)
	admin.GET("/inventory", api.AdminInventory)
	admin.GET("/orders", api.AdminOrders)
	admin.GET("/customers", api.AdminCustomers)
	admin.GET("/analytics", api.AdminAnalytics)

	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (api *API) Register(c *gin.Context) {
	var payload models.RegisterRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	payload.Email = utils.NormalizeEmail(payload.Email)
	var existing models.User
	if err := api.DB.Where("email = ?", payload.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	hash, err := utils.HashPassword(payload.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to hash password"})
		return
	}
	user := models.User{Name: payload.Name, Email: payload.Email, PasswordHash: hash, Role: models.RoleCustomer}
	if err := api.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create user"})
		return
	}
	accessToken, err := utils.GenerateAccessToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}
	c.JSON(http.StatusCreated, models.AuthResponse{User: user, AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "bearer", ExpiresIn: int64(api.Cfg.JWTAccessTTL / time.Second)})
}

func (api *API) Login(c *gin.Context) {
	var payload models.LoginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	payload.Email = utils.NormalizeEmail(payload.Email)
	var user models.User
	if err := api.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err := utils.CheckPassword(payload.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	accessToken, err := utils.GenerateAccessToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}
	c.JSON(http.StatusOK, models.AuthResponse{User: user, AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "bearer", ExpiresIn: int64(api.Cfg.JWTAccessTTL / time.Second)})
}

func (api *API) Refresh(c *gin.Context) {
	var payload models.RefreshRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	claims, err := utils.ParseToken(payload.RefreshToken, api.Cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}
	var user models.User
	if err := api.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	newAccess, err := utils.GenerateAccessToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token refresh failed"})
		return
	}
	newRefresh, err := utils.GenerateRefreshToken(user, api.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token refresh failed"})
		return
	}
	c.JSON(http.StatusOK, models.AuthResponse{User: user, AccessToken: newAccess, RefreshToken: newRefresh, TokenType: "bearer", ExpiresIn: int64(api.Cfg.JWTAccessTTL / time.Second)})
}

func (api *API) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (api *API) ListCategories(c *gin.Context) {
	var categories []models.Category
	if err := api.DB.Order("name asc").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (api *API) ListProducts(c *gin.Context) {
	query := api.DB.Preload("Category").Preload("Images").Where("status = ?", models.ProductStatusActive)
	categoryID := c.Query("category_id")
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	search := strings.TrimSpace(c.Query("search"))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(sku) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", like, like, like)
	}
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 12
	}
	offset := (page - 1) * limit
	var total int64
	query.Model(&models.Product{}).Count(&total)
	var products []models.Product
	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": products, "total": total, "page": page, "limit": limit})
}

func (api *API) GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := api.DB.Preload("Category").Preload("Images").Where("id = ? AND status = ?", id, models.ProductStatusActive).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (api *API) GetCart(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	cart := models.Cart{UserID: userID}
	if err := api.DB.Where("user_id = ?", userID).FirstOrCreate(&cart, models.Cart{UserID: userID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load cart"})
		return
	}
	var items []models.CartItem
	api.DB.Preload("Product").Where("cart_id = ?", cart.ID).Find(&items)
	cart.Items = items
	c.JSON(http.StatusOK, cart)
}

func (api *API) AddToCart(c *gin.Context) {
	var payload models.CartAddRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	userID := middleware.CurrentUserID(c)
	var cart models.Cart
	if err := api.DB.Where("user_id = ?", userID).FirstOrCreate(&cart, models.Cart{UserID: userID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create cart"})
		return
	}
	productID := uuid.MustParse(payload.ProductID)
	if productID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var product models.Product
	if err := api.DB.First(&product, "id = ? AND status = ?", productID, models.ProductStatusActive).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if product.StockQuantity <= 0 || int32(payload.Quantity) > product.StockQuantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient product stock"})
		return
	}
	var item models.CartItem
	if err := api.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&item).Error; err == nil {
		item.Quantity += payload.Quantity
		if int32(item.Quantity) > product.StockQuantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity exceeds available stock"})
			return
		}
		if err := api.DB.Save(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update cart"})
			return
		}
	} else {
		item = models.CartItem{CartID: cart.ID, ProductID: productID, Quantity: payload.Quantity}
		if err := api.DB.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item"})
			return
		}
	}
	c.JSON(http.StatusCreated, item)
}

func (api *API) UpdateCartItem(c *gin.Context) {
	var item models.CartItem
	id := c.Param("id")
	userID := middleware.CurrentUserID(c)
	if err := api.DB.Joins("JOIN carts ON carts.id = cart_items.cart_id").Where("cart_items.id = ? AND carts.user_id = ?", id, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}
	var payload struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity"})
		return
	}
	var product models.Product
	if err := api.DB.First(&product, "id = ?", item.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if int32(payload.Quantity) > product.StockQuantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity exceeds available stock"})
		return
	}
	item.Quantity = payload.Quantity
	if err := api.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update cart item"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (api *API) RemoveCartItem(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	id := c.Param("id")
	var item models.CartItem
	if err := api.DB.Joins("JOIN carts ON carts.id = cart_items.cart_id").Where("cart_items.id = ? AND carts.user_id = ?", id, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}
	if err := api.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "item removed from cart"})
}

func (api *API) Checkout(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	var payload models.CheckoutRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	var cart models.Cart
	if err := api.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}
	var items []models.CartItem
	if err := api.DB.Preload("Product").Where("cart_id = ?", cart.ID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart items"})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}
	if err := api.DB.Transaction(func(tx *gorm.DB) error {
		var subtotal decimal.Decimal
		for _, item := range items {
			if item.Product.Status != models.ProductStatusActive {
				return fmt.Errorf("product %s is unavailable", item.Product.Name)
			}
			if int32(item.Quantity) > item.Product.StockQuantity {
				return fmt.Errorf("insufficient stock for %s", item.Product.Name)
			}
			subtotal = subtotal.Add(item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		}
		tax := subtotal.Mul(decimal.NewFromFloat(0.08))
		shippingFee := decimal.NewFromFloat(12.00)
		total := subtotal.Add(tax).Add(shippingFee)
		orderNumber := utils.GenerateOrderNumber()
		order := models.Order{UserID: userID, OrderNumber: orderNumber, Status: models.OrderStatusPending, Subtotal: subtotal, Tax: tax, ShippingFee: shippingFee, Total: total, ShippingAddress: payload.ShippingAddress}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for _, item := range items {
			lineTotal := item.Product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			orderItem := models.OrderItem{OrderID: order.ID, ProductID: item.ProductID, ProductName: item.Product.Name, SKU: item.Product.SKU, UnitPrice: item.Product.Price, Quantity: item.Quantity, LineTotal: lineTotal}
			if err := tx.Create(&orderItem).Error; err != nil {
				return err
			}
			newQty := item.Product.StockQuantity - int32(item.Quantity)
			if newQty < 0 {
				return fmt.Errorf("insufficient stock for %s", item.Product.Name)
			}
			if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).Update("stock_quantity", newQty).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.InventoryMovement{ProductID: item.ProductID, ChangeQuantity: -int32(item.Quantity), PreviousQty: item.Product.StockQuantity, NewQty: newQty, Reason: "order checkout", AdminUserID: nil}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "order placed successfully"})
}

func (api *API) ListOrders(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	var orders []models.Order
	if err := api.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load orders"})
		return
	}
	for i := range orders {
		api.DB.Preload("Items").First(&orders[i], "id = ?", orders[i].ID)
	}
	c.JSON(http.StatusOK, orders)
}

func (api *API) GetOrder(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	id := c.Param("id")
	var order models.Order
	if err := api.DB.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	api.DB.Preload("Items").First(&order, "id = ?", order.ID)
	c.JSON(http.StatusOK, order)
}

func (api *API) CancelOrder(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	id := c.Param("id")
	var order models.Order
	if err := api.DB.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == models.OrderStatusCancelled || order.Status == models.OrderStatusDelivered || order.Status == models.OrderStatusShipped {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order cannot be cancelled in current state"})
		return
	}
	if err := api.DB.Model(&order).Update("status", models.OrderStatusCancelled).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}
	var items []models.OrderItem
	api.DB.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		api.DB.Model(&models.Product{}).Where("id = ?", item.ProductID).UpdateColumn("stock_quantity", gorm.Expr("stock_quantity + ?", item.Quantity))
		api.DB.Create(&models.InventoryMovement{ProductID: item.ProductID, ChangeQuantity: int32(item.Quantity), PreviousQty: 0, NewQty: 0, Reason: "order cancelled"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}

func (api *API) AdminOverview(c *gin.Context) {
	var totalProducts int64
	var lowStock int64
	var totalOrders int64
	var pendingOrders int64
	var sales decimal.Decimal
	api.DB.Model(&models.Product{}).Count(&totalProducts)
	api.DB.Model(&models.Product{}).Where("stock_quantity <= low_stock_threshold").Count(&lowStock)
	api.DB.Model(&models.Order{}).Count(&totalOrders)
	api.DB.Model(&models.Order{}).Where("status != ?", models.OrderStatusCancelled).Count(&pendingOrders)
	api.DB.Model(&models.Order{}).Select("COALESCE(SUM(total), 0)").Scan(&sales)
	c.JSON(http.StatusOK, gin.H{"total_products": totalProducts, "low_stock_products": lowStock, "total_orders": totalOrders, "pending_orders": pendingOrders, "sales": sales})
}

func (api *API) AdminListProducts(c *gin.Context) {
	var products []models.Product
	api.DB.Preload("Category").Preload("Images").Order("created_at desc").Find(&products)
	c.JSON(http.StatusOK, products)
}

func (api *API) AdminCreateProduct(c *gin.Context) {
	var payload models.CreateProductRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	parsedPrice, err := decimal.NewFromString(payload.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product price"})
		return
	}
	categoryID := utils.MustUUID(payload.CategoryID)
	if categoryID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	status := models.ProductStatusActive
	if payload.Status != "" {
		status = models.ProductStatus(payload.Status)
	}
	product := models.Product{Name: payload.Name, SKU: payload.SKU, Description: payload.Description, CategoryID: categoryID, Price: parsedPrice, StockQuantity: payload.StockQuantity, LowStockThreshold: payload.LowStockThreshold, Status: status}
	if err := api.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for idx, url := range payload.ImageURLs {
		if url == "" {
			continue
		}
		api.DB.Create(&models.ProductImage{ProductID: product.ID, ImageURL: url, SortOrder: idx})
	}
	c.JSON(http.StatusCreated, product)
}

func (api *API) AdminUpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := api.DB.First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	var payload models.UpdateProductRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if payload.Name != "" {
		product.Name = payload.Name
	}
	if payload.SKU != "" {
		product.SKU = payload.SKU
	}
	if payload.Description != "" {
		product.Description = payload.Description
	}
	if payload.CategoryID != "" {
		product.CategoryID = utils.MustUUID(payload.CategoryID)
	}
	if payload.Price != "" {
		parsed, err := decimal.NewFromString(payload.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product price"})
			return
		}
		product.Price = parsed
	}
	if payload.StockQuantity > 0 || payload.StockQuantity == 0 {
		product.StockQuantity = payload.StockQuantity
	}
	if payload.LowStockThreshold > 0 || payload.LowStockThreshold == 0 {
		product.LowStockThreshold = payload.LowStockThreshold
	}
	if payload.Status != "" {
		product.Status = models.ProductStatus(payload.Status)
	}
	if err := api.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}
	c.JSON(http.StatusOK, product)
}

func (api *API) AdminDeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := api.DB.Model(&models.Product{}).Where("id = ?", id).Update("status", models.ProductStatusArchived).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to archive product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product archived"})
}

func (api *API) AdminInventory(c *gin.Context) {
	var movements []models.InventoryMovement
	if err := api.DB.Order("created_at desc").Limit(100).Find(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load inventory movements"})
		return
	}
	c.JSON(http.StatusOK, movements)
}

func (api *API) AdminOrders(c *gin.Context) {
	var orders []models.Order
	api.DB.Preload("Items").Order("created_at desc").Find(&orders)
	c.JSON(http.StatusOK, orders)
}

func (api *API) AdminCustomers(c *gin.Context) {
	var users []models.User
	api.DB.Where("role = ?", models.RoleCustomer).Order("created_at desc").Find(&users)
	c.JSON(http.StatusOK, users)
}

func (api *API) AdminAnalytics(c *gin.Context) {
	ordersByStatus := map[string]int64{}
	for _, st := range []models.OrderStatus{models.OrderStatusPending, models.OrderStatusConfirmed, models.OrderStatusProcessing, models.OrderStatusShipped, models.OrderStatusDelivered, models.OrderStatusCancelled} {
		var count int64
		api.DB.Model(&models.Order{}).Where("status = ?", st).Count(&count)
		ordersByStatus[string(st)] = count
	}
	var monthlyRevenue []map[string]any
	api.DB.Raw("SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COALESCE(SUM(total), 0) AS revenue FROM orders GROUP BY TO_CHAR(created_at, 'YYYY-MM') ORDER BY month DESC LIMIT 6").Scan(&monthlyRevenue)
	c.JSON(http.StatusOK, gin.H{"orders_by_status": ordersByStatus, "monthly_revenue": monthlyRevenue})
}

func sanitizeStatus(input string) string {
	return strings.ToUpper(strings.TrimSpace(input))
}

func paginate(items []models.Product, page, limit int) []models.Product {
	if page <= 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	start := (page - 1) * limit
	if start >= len(items) {
		return nil
	}
	end := start + limit
	if end > len(items) {
		end = len(items)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items[start:end]
}
