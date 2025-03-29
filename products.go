package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/jung-kurt/gofpdf"
	"net/http"
	"strings"
	"time"
)

type ProductHandler struct {
	db *pg.DB
}

// @Summary      Get products
// @Description  Fetch products with pagination and filtering
// @Param        perPage  		query  int     false   "Number of products per page"
// @Param        field    		query  string  false   "Field to filter by (e.g., supplier, category)"
// @Param        values   		query  array   false   "Values of field"
// @Param        last_reference query  string  false   "The last reference of previous page"
// @Success      200  {array}  map[string]interface{}
// @Router       /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var req ProductRequest
	c.BindQuery(&req)

	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	/*
		 Parse dynamic filters:
			- field: field needed to query
			- values: values needed to query
			- example: reference = ["PROD-202401-029", "PROD-202401-039"]
	*/
	params := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key == "field" && len(values) > 0 {
			params[key] = values[0]
		}
		if key == "values" {
			params[key] = values
		}
	}

	products := make([]Product, 0)
	query := h.db.Model(&products)

	fieldToQuery := params["field"].(string)
	values := params["values"]
	if fieldToQuery != "" && values != nil {
		if fieldToQuery == "name" {
			fieldToQuery = "product.name"
		}
		if fieldToQuery == "category" {
			fieldToQuery = "category.name"
		}
		if fieldToQuery == "supplier" {
			fieldToQuery = "supplier.name"
		}
		query.Where(fmt.Sprintf("%v IN (?)", fieldToQuery), pg.In(values))
	}

	if req.LastReference != "" {
		query.Where("reference < ?", req.LastReference)
	}

	err := query.Relation("Category").Relation("Supplier").
		Order("reference DESC").
		Limit(req.PerPage).
		Select()

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}

// @Summary      Create product
// @Param        request  body  ProductCreateRequest  true  "Product filter request"
// @Success      200  {array}  map[string]interface{}
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req ProductCreateRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"msg":   "invalid request body",
		})
		return
	}

	_, err := h.db.Model(&Product{
		Name:       req.Name,
		Reference:  req.Reference,
		Status:     req.Status,
		CategoryID: req.CategoryID,
		Price:      req.Price,
		StockCity:  req.StockCity,
		SupplierID: req.SupplierID,
		Quantity:   req.Quantity,
	}).Insert()

	if err != nil {
		if strings.Contains(err.Error(), "products_category_id_fkey") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "category_id not exists",
			})
			return
		}

		if strings.Contains(err.Error(), "products_supplier_id_fkey") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "supplier_id not exists",
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when create product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "create product successfully",
	})
}

// @Summary      Get all categories of products
// @Success      200  {array}  map[string]interface{}
// @Router       /products/categories [get]
func (h *ProductHandler) GetCategories(c *gin.Context) {
	categories := make([]Category, 0)
	err := h.db.Model(&categories).Select()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// @Summary      Get all suppliers of products
// @Success      200  {array}  map[string]interface{}
// @Router       /products/suppliers [get]
func (h *ProductHandler) GetSuppliers(c *gin.Context) {
	suppliers := make([]Supplier, 0)
	err := h.db.Model(&suppliers).Select()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get suppliers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suppliers": suppliers,
	})
}

// @Summary      Update product
// @Param        request  body  ProductUpdateRequest  true  "Product filter request"
// @Param        id  path  int  true  "Product ID"
// @Success      200  {array}  map[string]interface{}
// @Router       /products/:id [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	var req ProductUpdateRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"msg":   "invalid request body",
		})
		return
	}

	id := c.Param("id")
	product := &Product{ID: id}
	if err := h.db.Model(product).WherePK().Select(); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "product not found",
		})
		return
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Reference != nil {
		product.Reference = *req.Reference
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.StockCity != nil {
		product.StockCity = *req.StockCity
	}
	if req.SupplierID != nil {
		product.SupplierID = *req.SupplierID
	}
	if req.Quantity != nil {
		product.Quantity = *req.Quantity
	}

	_, err := h.db.Model(product).WherePK().Update()
	if err != nil {
		if strings.Contains(err.Error(), "products_category_id_fkey") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "category_id not exists",
			})
			return
		}

		if strings.Contains(err.Error(), "products_supplier_id_fkey") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "supplier_id not exists",
			})
			return
		}

		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when update product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "update product successfully",
	})
}

// @Summary      Delete product
// @Param        id  path  int  true  "Product ID"
// @Success      200  {array}  map[string]interface{}
// @Router       /products/:id [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	product := &Product{ID: id}
	if err := h.db.Model(product).WherePK().Select(); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "product not found",
		})
		return
	}

	_, err := h.db.Model(product).WherePK().Delete()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when delete product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "delete product successfully",
	})
}

// @Summary      Statistics products per category
// @Success      200  {array}  map[string]interface{}
// @Router       /api/statistics/products-per-category [get]
func (h *ProductHandler) StatisticsProductsPerCategory(c *gin.Context) {
	rsp := make([]ProductsPerCategoryResponse, 0)

	err := h.db.Model(&Product{}).
		Join("JOIN categories ON categories.id = product.category_id").
		ColumnExpr("categories.name AS category_name, COUNT(*) AS total_products").
		Group("categories.name").Select(&rsp)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when statistic products by category",
		})
		return
	}

	c.JSON(http.StatusOK, rsp)
}

// @Summary      Statistics products per supplier
// @Success      200  {array}  map[string]interface{}
// @Router       /api/statistics/products-per-supplier [get]
func (h *ProductHandler) StatisticsProductsPerSupplier(c *gin.Context) {
	rsp := make([]ProductsPerSupplierResponse, 0)

	err := h.db.Model(&Product{}).
		Join("JOIN suppliers ON suppliers.id = product.supplier_id").
		ColumnExpr("suppliers.name AS supplier_name, COUNT(*) AS total_products").
		Group("suppliers.name").Select(&rsp)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when statistic products by supplier",
		})
		return
	}

	c.JSON(http.StatusOK, rsp)
}

// @Summary      Export products
// @Success      200 {file}  pdf
// @Router       /products/export [get]
func (h *ProductHandler) ExportProduct(c *gin.Context) {
	pdf := gofpdf.New("P", "mm", "A2", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, "Test - FE Data")
	pdf.Ln(10)

	header := []string{
		"Product Reference", "Product Name", "Date Added", "Status", "Product Category",
		"Price", "Stock Location (City)", "Supplier", "Available Quantity",
	}

	products := make([]Product, 0)
	query := h.db.Model(&products)

	/*
		 Parse dynamic filters:
			- field: field needed to query
			- values: values needed to query
			- example: reference = ["PROD-202401-029", "PROD-202401-039"]
	*/
	params := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key == "field" && len(values) > 0 {
			params[key] = values[0]
		}
		if key == "values" {
			params[key] = values
		}
	}

	fieldToQuery, _ := params["field"].(string)
	values := params["values"]
	if fieldToQuery != "" && values != nil {
		if fieldToQuery == "name" {
			fieldToQuery = "product.name"
		}
		if fieldToQuery == "category" {
			fieldToQuery = "category.name"
		}
		if fieldToQuery == "supplier" {
			fieldToQuery = "supplier.name"
		}
		query.Where(fmt.Sprintf("%v IN (?)", fieldToQuery), pg.In(values))
	}

	err := query.Relation("Category").Relation("Supplier").Order("reference DESC").Select()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get products",
		})
		return
	}

	data := make([][]string, 0)
	for _, product := range products {
		d := []string{product.Reference, product.Name, product.AddedDate.Format(time.DateOnly), product.Status}

		if product.Category != nil {
			d = append(d, product.Category.Name)
		} else {
			d = append(d, "")
		}

		d = append(d, fmt.Sprintf("%v", product.Price))
		d = append(d, product.StockCity)

		if product.Supplier != nil {
			d = append(d, product.Supplier.Name)
		} else {
			d = append(d, "")
		}

		d = append(d, fmt.Sprintf("%v", product.Quantity))

		data = append(data, d)
	}

	colWidths := []float64{45, 60, 30, 30, 50, 30, 50, 40, 50}

	// Draw header of table
	pdf.SetFont("Arial", "B", 12)
	for i, col := range header {
		pdf.CellFormat(colWidths[i], 10, col, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Draw data of table
	pdf.SetFont("Arial", "", 12)
	for _, row := range data {
		for i, col := range row {
			pdf.CellFormat(colWidths[i], 10, col, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	var pdfBuffer bytes.Buffer
	err = pdf.Output(&pdfBuffer)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when export products",
		})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=output.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBuffer.Bytes())
}

// @Summary      Get all cities of products
// @Success      200  {array}  map[string]interface{}
// @Router       /products/cities [get]
func (h *ProductHandler) GetCities(c *gin.Context) {
	cities := make([]string, 0)
	err := h.db.Model(&Product{}).Column("stock_city").Group("stock_city").Select(&cities)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "have error when get cities",
		})
		return
	}

	c.JSON(http.StatusOK, cities)
}
