package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/joho/godotenv"
	"github.com/umahmood/haversine"
	"manage-products/handlers"
	"manage-products/middlewares"
	"os"
	"time"
)

/*
TODO:   - This is only fake data about Latitude and Longitude of some city
  - We can add features about manage cities and their latitude, longitude later
*/
var cityCoordinates = map[string]haversine.Coord{
	"Paris":     {Lat: 48.8566, Lon: 2.3522},
	"Bordeaux":  {Lat: 44.8378, Lon: -0.5792},
	"Lyon":      {Lat: 45.7640, Lon: 4.8357},
	"Toulouse":  {Lat: 43.6047, Lon: 1.4442},
	"Marseille": {Lat: 43.2965, Lon: 5.3698},
}

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	POSTGRES_USER := os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD := os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_HOST := os.Getenv("POSTGRES_HOST")
	POSTGRES_PORT := os.Getenv("POSTGRES_PORT")

	opt, err := pg.ParseURL(
		fmt.Sprintf("postgres://%v:%v@%v:%v/backend_test",
			POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_HOST, POSTGRES_PORT,
		),
	)
	if err != nil {
		panic(err)
	}

	db := pg.Connect(opt)

	productHandler := ProductHandler{db: db}

	userHandler := handlers.UserHandler{DB: db}

	r.POST("users/sign-up", userHandler.SignUp)

	r.POST("users/sign-in", userHandler.SignIn)

	r.GET("products", middlewares.AuthenticateMiddleware, productHandler.GetProducts)

	r.GET("products/categories", middlewares.AuthenticateMiddleware, productHandler.GetCategories)

	r.GET("products/suppliers", middlewares.AuthenticateMiddleware, productHandler.GetSuppliers)

	r.POST("products", middlewares.AuthenticateMiddleware, productHandler.CreateProduct)

	r.PUT("products/:id", middlewares.AuthenticateMiddleware, productHandler.UpdateProduct)

	r.DELETE("products/:id", middlewares.AuthenticateMiddleware, productHandler.DeleteProduct)

	r.GET("api/statistics/products-per-category", middlewares.AuthenticateMiddleware, productHandler.StatisticsProductsPerCategory)

	r.GET("api/statistics/products-per-supplier", middlewares.AuthenticateMiddleware, productHandler.StatisticsProductsPerSupplier)

	r.GET("products/export", middlewares.AuthenticateMiddleware, productHandler.ExportProduct)

	r.GET("/distance", middlewares.AuthenticateMiddleware, calculateDistance)

	r.GET("products/cities", middlewares.AuthenticateMiddleware, productHandler.GetCities)

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

type ProductsPerCategoryResponse struct {
	CategoryName  string `json:"category_name"`
	TotalProducts int    `json:"total_products"`
}

type ProductsPerSupplierResponse struct {
	SupplierName  string `json:"supplier_name"`
	TotalProducts int    `json:"total_products"`
}
