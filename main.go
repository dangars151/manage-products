package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"net/http"
	"strings"
	"time"
)

func main() {
	r := gin.Default()

	// TODO: read user, password, host... from env
	opt, err := pg.ParseURL("postgres://postgres:M1sIWvQ2D4MfWke7ReSt2IFHVPRXtpp6@3.1.28.125:5432/backend_test")
	if err != nil {
		panic(err)
	}

	db := pg.Connect(opt)

	r.GET("products", func(c *gin.Context) {
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
		query := db.Model(&products)

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

		err = query.Relation("Category").Relation("Supplier").
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
	})

	r.GET("products/categories", func(c *gin.Context) {
		categories := make([]Category, 0)
		err := db.Model(&categories).Select()
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
	})

	r.GET("products/suppliers", func(c *gin.Context) {
		suppliers := make([]Supplier, 0)
		err := db.Model(&suppliers).Select()
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
	})

	r.POST("products", func(c *gin.Context) {
		var req ProductCreateRequest
		if err := c.Bind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
				"msg":   "invalid request body",
			})
			return
		}

		_, err = db.Model(&Product{
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
	})

	r.PUT("products/:id", func(c *gin.Context) {
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
		if err = db.Model(product).WherePK().Select(); err != nil {
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

		_, err = db.Model(product).WherePK().Update()
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
	})

	r.DELETE("products/:id", func(c *gin.Context) {
		id := c.Param("id")
		product := &Product{ID: id}
		if err = db.Model(product).WherePK().Select(); err != nil {
			fmt.Println(err)
			c.JSON(http.StatusNotFound, gin.H{
				"msg": "product not found",
			})
			return
		}

		_, err = db.Model(product).WherePK().Delete()
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
	})

	r.Run()
}

type Product struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Reference  string    `json:"reference"`
	AddedDate  time.Time `json:"added_date"`
	Status     string    `json:"status"`
	CategoryID string    `json:"category_id"`
	Price      float64   `json:"price"`
	StockCity  string    `json:"stock_city"`
	SupplierID string    `json:"supplier_id"`
	Quantity   int       `json:"quantity"`
	Category   *Category `json:"category" pg:"rel:has-one"`
	Supplier   *Supplier `json:"supplier" pg:"rel:has-one"`
}

type ProductCreateRequest struct {
	Name       string  `json:"name" binding:"required"`
	Reference  string  `json:"reference" binding:"required"`
	Status     string  `json:"status"`
	CategoryID string  `json:"category_id"`
	Price      float64 `json:"price"`
	StockCity  string  `json:"stock_city"`
	SupplierID string  `json:"supplier_id"`
	Quantity   int     `json:"quantity"`
}

type ProductUpdateRequest struct {
	Name       *string  `json:"name"`
	Reference  *string  `json:"reference"`
	Status     *string  `json:"status"`
	CategoryID *string  `json:"category_id"`
	Price      *float64 `json:"price"`
	StockCity  *string  `json:"stock_city"`
	SupplierID *string  `json:"supplier_id"`
	Quantity   *int     `json:"quantity"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductRequest struct {
	LastReference string `form:"last_reference"`
	PerPage       int    `form:"perPage"`
}

type Supplier struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
