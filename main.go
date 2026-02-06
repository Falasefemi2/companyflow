// @title CompanyFlow API
// @version 1.0
// @description API documentation for CompanyFlow services.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Provide a valid JWT token as: "Bearer <token>"
// @security BearerAuth
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/falasefemi2/companyflowlow/config"
	"github.com/falasefemi2/companyflowlow/database"
	_ "github.com/falasefemi2/companyflowlow/docs"
	"github.com/falasefemi2/companyflowlow/handlers"
	"github.com/falasefemi2/companyflowlow/repositories"
	"github.com/falasefemi2/companyflowlow/services"
)

func main() {
	fmt.Println("connecting to database")
	pool, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("Running migrations...")
	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	router := mux.NewRouter()

	// CORS Middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	router.Use(corsMiddleware)
	// Explicit OPTIONS handler so preflight requests don't get a 405 from mux.
	router.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	companyRepo := repositories.NewCompanyRepository(pool)
	companyService := services.NewCompanyService(companyRepo)
	companyHandler := handlers.NewCompanyHandler(companyService)
	companyHandler.RegisterRoutes(router)

	departmentRepo := repositories.NewDepartmentRepository(pool)
	departmentService := services.NewDepartmentService(departmentRepo)
	departmentHandler := handlers.NewDepartmentHandler(departmentService)
	departmentHandler.RegisterRoutes(router)

	employeeRepo := repositories.NewEmployeeRepository(pool)
	employeeService := services.NewEmployeeService(employeeRepo)
	employeeHandler := handlers.NewEmployeeHandler(employeeService)
	employeeHandler.RegisterRoutes(router)

	designationRepo := repositories.NewDesignationRepository(pool)
	designationService := services.NewDesignationService(designationRepo)
	designationHandler := handlers.NewDesignationHandler(designationService)
	designationHandler.RegisterRoutes(router)

	levelRepo := repositories.NewLevelRepository(pool)
	levelService := services.NewLevelService(levelRepo)
	levelHandler := handlers.NewLevelHandler(levelService)
	levelHandler.RegisterRoutes(router)

	roleRepo := repositories.NewRoleRepository(pool)
	roleService := services.NewRoleService(roleRepo)
	roleHandler := handlers.NewRoleHandler(roleService)
	roleHandler.RegisterRoutes(router)

	authService := services.NewAuthService(employeeRepo, companyRepo)
	authHandler := handlers.NewAuthHandler(authService)
	authHandler.RegisterRoutes(router)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	port := ":8080"
	fmt.Printf("\n✓ Server starting on http://localhost%s\n", port)
	fmt.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
