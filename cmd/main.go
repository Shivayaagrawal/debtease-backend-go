package main

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"

	// "DebtEase/internal/events"
	// "DebtEase/internal/handlers"
	"DebtEase/internal/logger"
	"DebtEase/internal/middleware"

	// "DebtEase/internal/notifications"
	"DebtEase/internal/pubsub"
	webapp_handlers "DebtEase/internal/webapp_specifics/handlers"

	// "DebtEase/internal/worker"
	// "context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	// "time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		logger.Logger.Warnw("Error loading .env file", "error", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Logger.Fatal("DB_URL not set in environment")
	}

	// jwtSecret := os.Getenv("JWT_SECRET")
	// if jwtSecret == "" {
	// 	logger.Logger.Fatal("JWT_SECRET is required in .env")
	// }

	platform := os.Getenv("PLATFORM")

	// Open DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Logger.Fatalw("Failed to open database", "error", err)
	}
	logger.Logger.Info("Database connection successful")
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Logger.Fatalw("Database ping failed", "error", err)
	}

	dbQueries := database.New(db)

	// rabbitmqURL := os.Getenv("RABBITMQ_URL")
	// if rabbitmqURL == "" {
	// 	rabbitmqURL = "amqp://guest:guest@localhost:5672/" // Default for local dev
	// 	logger.Logger.Warnw("RABBITMQ_URL not set, using default", "url", rabbitmqURL)
	// }

	// // Initialize RabbitMQ connection (lazy, will connect on first use)
	// // Test connection
	// conn, err := pubsub.GetConnection(rabbitmqURL)
	// if err != nil {
	// 	logger.Logger.Warnw("Failed to connect to RabbitMQ (notifications may not work)", "error", err)
	// 	// Don't fail startup - notifications are non-critical
	// } else {
	// 	logger.Logger.Info("RabbitMQ connection initialized")

	// 	// Declare exchange once (idempotent)
	// 	ch, err := pubsub.DeclareExchangeWithConnection(conn, events.ExchangeNotifications)
	// 	if err != nil {
	// 		logger.Logger.Warnw("Failed to declare exchange", "error", err)
	// 	} else {
	// 		ch.Close() // Close after declaration
	// 	}
	// 	// Start notification consumer
	// 	ctx := context.Background()
	// 	if err := notifications.StartConsumer(ctx, conn, dbQueries); err != nil {
	// 		logger.Logger.Warnw("Failed to start notification consumer (notifications may not work)", "error", err)
	// 		// Don't fail startup - notifications are non-critical
	// 	} else {
	// 		logger.Logger.Info("Notification consumer started successfully")
	// 	}
	// }

	// Initialize config
	cfg := &api.Config{
		DB:          dbQueries,
		Platform:    platform,
		// JWTSecret:   jwtSecret,
		// RabbitMQURL: rabbitmqURL,
	}

	// Setup router
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infow("Health check", "path", r.URL.Path, "method", r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static files with logging middleware
	// fileServer := http.FileServer(http.Dir("."))
	// mux.Handle("/app/", middleware.Logging(middleware.MetricsInc(cfg, http.StripPrefix("/app", fileServer))))

	// Admin
	// mux.HandleFunc("GET /admin/metrics", handlers.HandleMetrics(cfg))
	getAnalytics := webapp_handlers.HandleGetAnalytics(cfg)
	mux.Handle("GET /api/admin/analytics", middleware.Logging( getAnalytics))

	// // auth
	// mux.HandleFunc("POST /api/login", handlers.HandleLogin(cfg))
	// mux.HandleFunc("POST /api/register", handlers.HandleCreateUser(cfg))
	// mux.HandleFunc("POST /api/refresh", handlers.HandleTokenRefresh(cfg))
	// mux.HandleFunc("PUT /api/users", handlers.HandleUpdateUser(cfg))

	// // dashboard (authenticated)
	// getDashboard := handlers.HandleGetDashboard(cfg)
	// mux.Handle("GET /api/dashboard", middleware.Logging(middleware.AuthRequired(cfg, getDashboard)))

	// // debts (authenticated)
	// {
	// 	// POST /api/debts
	// 	createDebt := handlers.HandleCreateDebt(cfg)
	// 	mux.Handle("POST /api/debts", middleware.Logging(middleware.AuthRequired(cfg, createDebt)))

	// 	// GET /api/debts
	// 	listDebts := handlers.HandleListDebts(cfg)
	// 	mux.Handle("GET /api/debts", middleware.Logging(middleware.AuthRequired(cfg, listDebts)))

	// 	// GET /api/debts/{id}
	// 	getDebt := handlers.HandleGetDebt(cfg)
	// 	mux.Handle("GET /api/debts/{id}", middleware.Logging(middleware.AuthRequired(cfg, getDebt)))

	// 	// PATCH /api/debts/{id}/balance
	// 	updateBalance := handlers.HandleUpdateDebtBalance(cfg)
	// 	mux.Handle("PATCH /api/debts/{id}/balance", middleware.Logging(middleware.AuthRequired(cfg, updateBalance)))

	// 	// PATCH /api/debts/{id}/health
	// 	updateHealth := handlers.HandleUpdateDebtHealth(cfg)
	// 	mux.Handle("PATCH /api/debts/{id}/health", middleware.Logging(middleware.AuthRequired(cfg, updateHealth)))

	// 	// POST /api/debts/{id}/overdue
	// 	markOverdue := handlers.HandleMarkDebtOverdue(cfg)
	// 	mux.Handle("POST /api/debts/{id}/overdue", middleware.Logging(middleware.AuthRequired(cfg, markOverdue)))

	// 	// payments
	// 	recordPayment := handlers.HandleRecordPayment(cfg, db)
	// 	mux.Handle("POST /api/debts/{id}/payments", middleware.Logging(middleware.AuthRequired(cfg, recordPayment)))
	// 	listPayments := handlers.HandleListPayments(cfg)
	// 	mux.Handle("GET /api/debts/{id}/payments", middleware.Logging(middleware.AuthRequired(cfg, listPayments)))

	// 	// // DELETE /api/debts/{id}
	// 	// deleteDebt := handlers.HandleDeleteDebt(cfg)
	// 	// mux.Handle("DELETE /api/debts/{id}", middleware.Logging(middleware.AuthRequired(cfg, deleteDebt)))
	// }

	// Webapp calculator endpoints (no auth required)
	{
		baselineCalculate := webapp_handlers.HandleBaselineCalculate(cfg)
		mux.Handle("POST /api/v1/calculate/baseline", middleware.Logging(baselineCalculate))

		strategyCalculate := webapp_handlers.HandleStrategyCalculate(cfg)
		mux.Handle("POST /api/v1/calculate/strategy", middleware.Logging(strategyCalculate))

		createEvent := webapp_handlers.HandleCreateEvent(cfg)
		mux.Handle("POST /api/events", middleware.Logging(createEvent))

		pdfRequest := webapp_handlers.HandlePdfRequest(cfg)
		mux.Handle("POST /api/pdf-request", middleware.Logging(pdfRequest))

		waitlist := webapp_handlers.HandleWaitlist(cfg)
		mux.Handle("POST /api/waitlist", middleware.Logging(waitlist))
	}

	// // Start workers
	// ctx := context.Background()
	// worker.StartDailyAccrual(ctx, dbQueries, 24*time.Hour)              // Daily interest accrual for credit cards
	// worker.StartCreditCardStatementWorker(ctx, dbQueries, 24*time.Hour) // Credit card statement generation
	// worker.StartLoanEMIWorker(ctx, dbQueries, 24*time.Hour)             // Loan EMI processing

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Logger.Info("Shutting down gracefully...")
		pubsub.CloseConnection() // Close RabbitMQ connections
		os.Exit(0)
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local dev
	}

	// Start server
	logger.Logger.Infow("Server starting", "port", port, "platform", platform)
	logger.Logger.Fatal(http.ListenAndServe(":"+port, mux))
}
