package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/atam/atamlink/internal/config"
	"github.com/atam/atamlink/internal/handler"
	"github.com/atam/atamlink/internal/middleware"
	auditRepo "github.com/atam/atamlink/internal/mod_audit/repository"
	businessRepo "github.com/atam/atamlink/internal/mod_business/repository"
	"github.com/atam/atamlink/internal/mod_business/usecase"
	catalogRepo "github.com/atam/atamlink/internal/mod_catalog/repository"
	catalogUC "github.com/atam/atamlink/internal/mod_catalog/usecase"
	masterRepo "github.com/atam/atamlink/internal/mod_master/repository"
	masterUC "github.com/atam/atamlink/internal/mod_master/usecase"
	userRepo "github.com/atam/atamlink/internal/mod_user/repository"
	userUC "github.com/atam/atamlink/internal/mod_user/usecase"	
	"github.com/atam/atamlink/internal/service"
	"github.com/atam/atamlink/pkg/database"
	"github.com/atam/atamlink/pkg/logger"
	"github.com/atam/atamlink/pkg/utils"
)

// App adalah struct utama yang menampung semua komponen aplikasi.
type App struct {
	Config *config.Config
	Server *http.Server
	Log    logger.Logger
	DB     *sql.DB
	AuditService service.AuditService
}

// New membuat dan mengonfigurasi instance aplikasi baru.
func New() (*App, error) {
	// Muat environment variables dari .env, abaikan jika tidak ada.
	_ = godotenv.Load()

	// Inisialisasi komponen dasar
	cfg := config.Load()
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Inisialisasi semua dependensi (DI Container)
	// Services
	validator := utils.NewValidator()
	slugService := service.NewSlugService()
	uploadService := service.NewUploadService(cfg.Upload)
	
	// Repositories
	userRepository := userRepo.NewUserRepository(db)
	businessRepository := businessRepo.NewBusinessRepository(db)
	catalogRepository := catalogRepo.NewCatalogRepository(db)
	masterRepository := masterRepo.NewMasterRepository(db)
	auditRepository := auditRepo.NewAuditRepository(db)

	// Start audit service
	auditService := service.NewAuditService(auditRepository, log)
	auditService.Start()

	// Use Cases
	businessUseCase := usecase.NewBusinessUseCase(db, businessRepository, userRepository, slugService)
	catalogUseCase := catalogUC.NewCatalogUseCase(db, catalogRepository, businessRepository, slugService)
	masterUseCase := masterUC.NewMasterUseCase(db, masterRepository)
	userUseCase := userUC.NewUserUseCase(db, userRepository)

	// Handlers
	healthHandler := handler.NewHealthHandler(db)
	businessHandler := handler.NewBusinessHandler(businessUseCase, validator)
	catalogHandler := handler.NewCatalogHandler(catalogUseCase, uploadService, validator)
	masterHandler := handler.NewMasterHandler(masterUseCase, validator)
	userHandler := handler.NewUserHandler(userUseCase, validator)

	// Inisialisasi router Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Pasang middleware global
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS(cfg.CORS))

	// Daftarkan semua rute
	setupRoutes(router, cfg, auditService, healthHandler, businessHandler, catalogHandler, masterHandler, userHandler)

	// Konfigurasi server HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &App{
		Config: cfg,
		Server: srv,
		Log:    log,
		DB:     db,
		AuditService: auditService,
	}, nil
}

// Run memulai server HTTP dan menangani graceful shutdown.
func (a *App) Run() {
	// Jalankan server di goroutine terpisah agar tidak memblokir
	go func() {
		a.Log.Info("Server starting", logger.String("address", a.Server.Addr))
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Log.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	// Tunggu sinyal interupsi (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	a.Log.Info("Shutting down server...")

	// Stop audit service  
	a.AuditService.Stop()

	// Beri waktu 5 detik untuk menyelesaikan request yang sedang berjalan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Matikan koneksi database dan logger dengan rapi
	defer a.Log.Sync()
	defer a.DB.Close()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Log.Fatal("Server forced to shutdown", logger.Error(err))
	}

	a.Log.Info("Server exited properly")
}

// setupRoutes mendaftarkan semua handler ke router Gin.
func setupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	auditService service.AuditService,
	healthHandler *handler.HealthHandler,
	businessHandler *handler.BusinessHandler,
	catalogHandler *handler.CatalogHandler,
	masterHandler *handler.MasterHandler,
	userHandler *handler.UserHandler,
) {
	// Rute Health check (tidak perlu otentikasi)
	router.GET("/health", healthHandler.Check)
	router.GET("/health/db", healthHandler.CheckDB)

	// Rute untuk file statis (uploads)
	router.Static("/uploads", "./uploads")

	// Grup untuk semua rute API v1
	api := router.Group(cfg.API.Prefix)
	{
		// Terapkan middleware otentikasi
		if cfg.Auth.Bypass {
			api.Use(middleware.AuthBypass(cfg.Auth.BypassUserID, cfg.Auth.BypassProfileID))
		} else {
			api.Use(middleware.Auth())
		}

		// Audit middleware
		api.Use(middleware.Audit(auditService, nil))

		// Rute untuk modul Business
		businesses := api.Group("/businesses")
		{
			businesses.POST("", businessHandler.Create)
			businesses.GET("", businessHandler.List)
			businesses.GET("/:id", businessHandler.GetByID)
			businesses.PUT("/:id", businessHandler.Update)
			businesses.DELETE("/:id", businessHandler.Delete)
			// TODO: Tambahkan rute untuk user management di dalam business
		}

		// Rute untuk modul Catalog
		catalogs := api.Group("/catalogs")
		{
			catalogs.POST("", catalogHandler.Create)
			catalogs.GET("", catalogHandler.List)
			catalogs.GET("/:id", catalogHandler.GetByID)
			catalogs.PUT("/:id", catalogHandler.Update)
			catalogs.DELETE("/:id", catalogHandler.Delete)
			// TODO: Tambahkan rute untuk section dan card management
		}

		// Rute untuk modul Master Data
		// masters := api.Group("/masters")
		// {
		// 	masters.GET("/plans", masterHandler.ListPlans)
		// 	masters.GET("/themes", masterHandler.ListThemes)
		// }

		// Rute untuk modul Master Data
		masters := api.Group("/masters")
		{
			// Plans CRUD
			masters.POST("/plans", masterHandler.CreatePlan)
			masters.GET("/plans", masterHandler.ListPlans)
			masters.GET("/plans/:id", masterHandler.GetPlanByID)
			masters.PUT("/plans/:id", masterHandler.UpdatePlan)
			masters.DELETE("/plans/:id", masterHandler.DeletePlan)

			// Themes CRUD
			masters.POST("/themes", masterHandler.CreateTheme)
			masters.GET("/themes", masterHandler.ListThemes)
			masters.GET("/themes/:id", masterHandler.GetThemeByID)
			masters.PUT("/themes/:id", masterHandler.UpdateTheme)
			masters.DELETE("/themes/:id", masterHandler.DeleteTheme)
		}

		profile := api.Group("/profile")
		{
			profile.GET("", userHandler.GetProfile)
			profile.POST("", userHandler.CreateProfile)
			profile.PUT("", userHandler.UpdateProfile)
			profile.DELETE("", userHandler.DeleteProfile)
		}

		// Rute untuk User Management (admin)
		users := api.Group("/users")
		{
			users.GET("/profiles/:id", userHandler.GetProfileByID)
			users.PUT("/profiles/:id", userHandler.UpdateProfileByID)
			users.DELETE("/profiles/:id", userHandler.DeleteProfileByID)
		}
	}
}