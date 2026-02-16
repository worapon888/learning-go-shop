package server

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joefazee/learning-go-shop/internal/dto"
	"github.com/joefazee/learning-go-shop/internal/utils"
)

// ================== CATEGORY ==================

func (s *Server) createCategory(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	category, err := s.productService.CreateCategory(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create category", err)
		return
	}

	utils.CreatedResponse(c, "Category created successfully", category)
}

func (s *Server) getCategories(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	categories, err := s.productService.GetCategories()
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch categories", err)
		return
	}

	utils.SuccessResponse(c, "Categories retrieved successfully", categories)
}

func (s *Server) updateCategory(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID", err)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	category, err := s.productService.UpdateCategory(id, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update category", err)
		return
	}

	utils.SuccessResponse(c, "Category updated successfully", category)
}

func (s *Server) deleteCategory(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid category ID", err)
		return
	}

	if err := s.productService.DeleteCategory(id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete category", err)
		return
	}

	utils.SuccessResponse(c, "Category deleted successfully", nil)
}

// ================== PRODUCT ==================

func (s *Server) createProduct(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	product, err := s.productService.CreateProduct(&req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to create product", err)
		return
	}

	utils.CreatedResponse(c, "Product created successfully", product)
}

func (s *Server) getProducts(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	page := parseIntQuery(c, "page", 1, 1, 1_000_000)
	limit := parseIntQuery(c, "limit", 10, 1, 1000)

	products, meta, err := s.productService.GetProducts(page, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to fetch products", err)
		return
	}

	// meta เป็น pointer => กัน nil (บาง service อาจส่ง nil)
	if meta == nil {
		utils.SuccessResponse(c, "Products retrieved successfully", products)
		return
	}

	utils.PaginatedSuccessResponse(c, "Products retrieved successfully", products, *meta)
}

func (s *Server) getProduct(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	product, err := s.productService.GetProduct(id)
	if err != nil {
		utils.NotFoundResponse(c, "Product not found")
		return
	}

	utils.SuccessResponse(c, "Product retrieved successfully", product)
}

func (s *Server) updateProduct(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	product, err := s.productService.UpdateProduct(id, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update product", err)
		return
	}

	utils.SuccessResponse(c, "Product updated successfully", product)
}

func (s *Server) deleteProduct(c *gin.Context) {
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	if err := s.productService.DeleteProduct(id); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to delete product", err)
		return
	}

	utils.SuccessResponse(c, "Product deleted successfully", nil)
}

// ================== PRODUCT IMAGE UPLOAD ==================
// ต้องมี s.uploadService (UploadService) + productService.AddProductImage(...) อยู่จริง

func (s *Server) uploadProductImage(c *gin.Context) {
	if s.uploadService == nil {
		utils.InternalServerErrorResponse(c, "uploadService not initialized", nil)
		return
	}
	if s.productService == nil {
		utils.InternalServerErrorResponse(c, "productService not initialized", nil)
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.BadRequestResponse(c, "Invalid product ID", err)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded", err)
		return
	}

	url, err := s.uploadService.UploadProductImage(id, file)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to upload image", err)
		return
	}

	if err := s.productService.AddProductImage(id, url, file.Filename); err != nil {
		utils.InternalServerErrorResponse(c, "Failed to save image record", err)
		return
	}

	utils.SuccessResponse(c, "Image uploaded successfully", gin.H{"url": url})
}

// ================== helpers ==================

func parseUintParam(c *gin.Context, key string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(key), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func parseIntQuery(c *gin.Context, key string, def, min, max int) int {
	vStr := c.DefaultQuery(key, strconv.Itoa(def))
	v, err := strconv.Atoi(vStr)
	if err != nil {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
