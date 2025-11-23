package main

import (
	"fmt"
	"net/http"
	"os"
	"quokkaq-go-backend/internal/config"
	"quokkaq-go-backend/internal/handlers"
	"quokkaq-go-backend/internal/jobs"
	authmiddleware "quokkaq-go-backend/internal/middleware"
	"quokkaq-go-backend/internal/models"
	"quokkaq-go-backend/internal/repository"
	"quokkaq-go-backend/internal/services"
	"quokkaq-go-backend/internal/ws"
	"quokkaq-go-backend/pkg/database"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// @title           QuokkaQ Go Backend API
// @version         1.0
// @description     This is the backend API for QuokkaQ, rewritten in Go.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.swagger.io/support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:3001
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.Load()
	database.Connect()

	// Auto Migrate
	database.AutoMigrate(
		&models.Company{},
		&models.Unit{},
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.UserUnit{},
		&models.Service{},
		&models.Counter{},
		&models.Ticket{},
		&models.TicketHistory{},
		&models.TicketNumberSequence{},
		&models.Booking{},
		&models.Notification{},
		&models.AuditLog{},
		&models.UnitMaterial{},
		&models.Invitation{},
		&models.MessageTemplate{},
		&models.PasswordResetToken{},
	)

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// Background Jobs
	jobClient := jobs.NewJobClient()
	defer jobClient.Close()

	// Initialize Storage and TTS Services
	storageService := services.NewStorageService()
	ttsService := services.NewTtsService(storageService)

	jobWorker := jobs.NewJobWorker(ttsService)
	jobWorker.Start()
	defer jobWorker.Stop()

	// Repositories
	userRepo := repository.NewUserRepository()
	unitRepo := repository.NewUnitRepository()
	ticketRepo := repository.NewTicketRepository()
	serviceRepo := repository.NewServiceRepository()
	counterRepo := repository.NewCounterRepository()
	bookingRepo := repository.NewBookingRepository()
	templateRepo := repository.NewTemplateRepository()
	invitationRepo := repository.NewInvitationRepository()

	// Services
	userService := services.NewUserService(userRepo)
	mailService := services.NewMailService()
	authService := services.NewAuthService(userRepo, mailService)
	unitService := services.NewUnitService(unitRepo)
	ticketService := services.NewTicketService(ticketRepo, counterRepo, serviceRepo, hub, jobClient)
	serviceService := services.NewServiceService(serviceRepo)
	counterService := services.NewCounterService(counterRepo, ticketRepo, userRepo)
	bookingService := services.NewBookingService(bookingRepo)
	shiftService := services.NewShiftService(ticketRepo, counterRepo)
	templateService := services.NewTemplateService(templateRepo)
	invitationService := services.NewInvitationService(invitationRepo, mailService, userRepo)

	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(authService)
	unitHandler := handlers.NewUnitHandler(unitService, storageService)
	ticketHandler := handlers.NewTicketHandler(ticketService)
	serviceHandler := handlers.NewServiceHandler(serviceService)
	counterHandler := handlers.NewCounterHandler(counterService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	shiftHandler := handlers.NewShiftHandler(shiftService)
	templateHandler := handlers.NewTemplateHandler(templateService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "https://quokkaq.v-b.tech", "https://app.quokkaq.v-b.tech"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from QuokkaQ Go Backend!"))
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		// Read the swagger.json file
		content, err := os.ReadFile("./docs/swagger.json")
		if err != nil {
			http.Error(w, "Failed to read swagger.json", http.StatusInternalServerError)
			return
		}

		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecContent: string(content),
			CustomOptions: scalar.CustomOptions{
				PageTitle: "QuokkaQ API Reference",
			},
			DarkMode: true,
		})

		if err != nil {
			fmt.Printf("%v", err)
		}

		fmt.Fprintln(w, htmlContent)
	})

	// Serve the swagger.json file
	r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/forgot-password", authHandler.RequestPasswordReset)
		r.Post("/reset-password", authHandler.ResetPassword)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.JWTAuth)
			r.Get("/me", authHandler.GetMe)
		})
	})

	r.Route("/system", func(r chi.Router) {
		r.Get("/status", userHandler.GetSystemStatus)
		r.Post("/setup", userHandler.SetupFirstAdmin)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.CreateUser)
		r.Get("/", userHandler.GetAllUsers)
		r.Get("/{id}", userHandler.GetUserByID)
		r.Patch("/{id}", userHandler.UpdateUser)
		r.Delete("/{id}", userHandler.DeleteUser)

		// User-Unit operations
		r.Get("/{id}/units", userHandler.GetUserUnits)
		r.Post("/{id}/units/assign", userHandler.AssignUnit)
		r.Post("/{id}/units/remove", userHandler.RemoveUnit)
	})

	r.Route("/units", func(r chi.Router) {
		r.Post("/", unitHandler.CreateUnit)
		r.Get("/", unitHandler.GetAllUnits)
		r.Get("/{id}", unitHandler.GetUnitByID)
		r.Patch("/{id}", unitHandler.UpdateUnit)
		r.Delete("/{id}", unitHandler.DeleteUnit)

		// Unit-specific routes
		r.Post("/{unitId}/tickets", ticketHandler.CreateTicket)
		r.Get("/{unitId}/tickets", ticketHandler.GetTicketsByUnit)
		r.Post("/{unitId}/call-next", ticketHandler.CallNext)
		r.Get("/{unitId}/services", serviceHandler.GetServicesByUnit)
		r.Get("/{unitId}/services-tree", serviceHandler.GetServicesByUnit) // Alias for tree view
		r.Get("/{unitId}/counters", counterHandler.GetCountersByUnit)
		r.Post("/{unitId}/counters", counterHandler.CreateCounter)
		r.Get("/{unitId}/bookings", bookingHandler.GetBookingsByUnit)
		r.Get("/{unitId}/shift/dashboard", shiftHandler.GetDashboardStats)
		r.Get("/{unitId}/shift/queue", shiftHandler.GetQueueTickets)
		r.Get("/{unitId}/shift/counters", shiftHandler.GetShiftCounters)
		r.Post("/{unitId}/shift/eod", shiftHandler.ExecuteEndOfDay)

		// Unit Materials
		r.Post("/{unitId}/materials", unitHandler.AddMaterial)
		r.Get("/{unitId}/materials", unitHandler.GetMaterials)
		r.Delete("/{unitId}/materials/{materialId}", unitHandler.DeleteMaterial)

		// Ad Settings
		r.Patch("/{unitId}/ad-settings", unitHandler.UpdateAdSettings)
	})

	r.Route("/services", func(r chi.Router) {
		r.Post("/", serviceHandler.CreateService)
		r.Get("/{id}", serviceHandler.GetServiceByID)
		r.Put("/{id}", serviceHandler.UpdateService)
		r.Delete("/{id}", serviceHandler.DeleteService)
	})

	r.Route("/counters", func(r chi.Router) {
		// Public routes (if any)
		r.Get("/{id}", counterHandler.GetCounterByID)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(authmiddleware.JWTAuth)
			r.Post("/", counterHandler.CreateCounter)
			r.Put("/{id}", counterHandler.UpdateCounter)
			r.Delete("/{id}", counterHandler.DeleteCounter)
			r.Post("/{id}/occupy", counterHandler.Occupy)
			r.Post("/{id}/release", counterHandler.Release)
			r.Post("/{id}/force-release", counterHandler.ForceRelease)
			r.Post("/{id}/call-next", counterHandler.CallNext)
		})
	})

	r.Route("/bookings", func(r chi.Router) {
		r.Post("/", bookingHandler.CreateBooking)
		r.Get("/{id}", bookingHandler.GetBookingByID)
		r.Put("/{id}", bookingHandler.UpdateBooking)
		r.Delete("/{id}", bookingHandler.DeleteBooking)
	})

	r.Route("/templates", func(r chi.Router) {
		r.Post("/", templateHandler.CreateTemplate)
		r.Get("/", templateHandler.GetAllTemplates)
		r.Get("/{id}", templateHandler.GetTemplateByID)
		r.Put("/{id}", templateHandler.UpdateTemplate)
		r.Delete("/{id}", templateHandler.DeleteTemplate)
	})

	r.Route("/invitations", func(r chi.Router) {
		r.Post("/", invitationHandler.CreateInvitation)
		r.Get("/", invitationHandler.GetAllInvitations)
		r.Delete("/{id}", invitationHandler.DeleteInvitation)
		r.Patch("/{id}/resend", invitationHandler.ResendInvitation)
		r.Get("/token/{token}", invitationHandler.GetInvitationByToken)
		r.Post("/register", invitationHandler.RegisterUser)
	})

	uploadHandler := handlers.NewUploadHandler(storageService)
	r.Post("/upload", uploadHandler.UploadLogo)

	r.Route("/tickets", func(r chi.Router) {
		r.Get("/{id}", ticketHandler.GetTicketByID)
		r.Patch("/{id}/status", ticketHandler.UpdateStatus)
		r.Post("/{id}/recall", ticketHandler.Recall)
		r.Post("/{id}/pick", ticketHandler.Pick)
		r.Post("/{id}/transfer", ticketHandler.Transfer)
		r.Post("/{id}/return", ticketHandler.ReturnToQueue)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Server starting on port %s\n", port)
	http.ListenAndServe(":"+port, r)
}
