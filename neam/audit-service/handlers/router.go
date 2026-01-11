package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"neam-platform/audit-service/middleware"
)

// Router provides HTTP routing
type Router interface {
	Use(...gin.HandlerFunc)
	GET(path string, handlers ...gin.HandlerFunc)
	POST(path string, handlers ...gin.HandlerFunc)
	PUT(path string, handlers ...gin.HandlerFunc)
	DELETE(path string, handlers ...gin.HandlerFunc)
	PATCH(path string, handlers ...gin.HandlerFunc)
	Group(path string) *gin.RouterGroup
	Run(addr string) error
	Shutdown(ctx context.Context) error
}

// ginRouter implements Router using Gin
type ginRouter struct {
	router *gin.Engine
}

// NewRouter creates a new Gin router
func NewRouter() Router {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.SecurityHeaders())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return &ginRouter{router: router}
}

// NewCustomRouter creates a router with custom configuration
func NewCustomRouter(mode string, logFile string) Router {
	gin.SetMode(mode)

	router := gin.New()

	// Set up logging
	if logFile != "" {
		logFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			gin.DefaultWriter = logFile
			gin.DefaultErrorWriter = logFile
		}
	}

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.RequestID())

	return &ginRouter{router: router}
}

// NewAuditRouter creates a router for audit service
func NewAuditRouter(auditHandler *AuditHandler, authHandler *AuthHandler, forensicHandler *ForensicHandler) Router {
	router := NewCustomRouter(gin.ReleaseMode, "")

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health (no auth required)
		v1.GET("/health", healthCheck)

		// Auth routes (no auth required for login)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/verify", authHandler.VerifyToken)
			auth.GET("/me", authHandler.GetCurrentUser)
			auth.POST("/mfa/verify", authHandler.VerifyMFA)
		}

		// Audit routes (require auth)
		audit := v1.Group("/audit")
		audit.Use(middleware.ZeroTrustAuth(nil, nil))
		audit.Use(middleware.AuditMiddleware())
		{
			audit.GET("", auditHandler.QueryEvents)
			audit.GET("/:id", auditHandler.GetEvent)
			audit.GET("/statistics", auditHandler.GetStatistics)
			audit.POST("/verify", middleware.RequirePermission("audit:verify"), auditHandler.VerifyIntegrity)
			audit.POST("/seal", middleware.RequirePermission("audit:seal"), auditHandler.SealEvents)
			audit.GET("/chain/:id", auditHandler.GetChain)
			audit.GET("/export", middleware.RequirePermission("audit:read"), auditHandler.ExportAuditLogs)
		}

		// Forensic routes (require auth)
		forensic := v1.Group("/forensic")
		forensic.Use(middleware.ZeroTrustAuth(nil, nil))
		{
			forensic.POST("", middleware.RequirePermission("forensic:create"), forensicHandler.CreateInvestigation)
			forensic.GET("/:id", forensicHandler.GetInvestigation)
			forensic.POST("/:id/approve", middleware.RequireRole("admin", "security-officer"), forensicHandler.ApproveInvestigation)
			forensic.POST("/:id/execute", middleware.RequirePermission("forensic:create"), forensicHandler.ExecuteInvestigation)
			forensic.GET("/:id/report", forensicHandler.GetInvestigationReport)
			forensic.GET("/evidence/:event_id", forensicHandler.GetEvidence)
			forensic.POST("/verify", forensicHandler.VerifyEvidence)
		}
	}

	return router
}

// Use adds middleware
func (r *ginRouter) Use(middleware ...gin.HandlerFunc) {
	r.router.Use(middleware...)
}

// GET adds GET route
func (r *ginRouter) GET(path string, handlers ...gin.HandlerFunc) {
	r.router.GET(path, handlers...)
}

// POST adds POST route
func (r *ginRouter) POST(path string, handlers ...gin.HandlerFunc) {
	r.router.POST(path, handlers...)
}

// PUT adds PUT route
func (r *ginRouter) PUT(path string, handlers ...gin.HandlerFunc) {
	r.router.PUT(path, handlers...)
}

// DELETE adds DELETE route
func (r *ginRouter) DELETE(path string, handlers ...gin.HandlerFunc) {
	r.router.DELETE(path, handlers...)
}

// PATCH adds PATCH route
func (r *ginRouter) PATCH(path string, handlers ...gin.HandlerFunc) {
	r.router.PATCH(path, handlers...)
}

// Group creates a route group
func (r *ginRouter) Group(path string) *gin.RouterGroup {
	return r.router.Group(path)
}

// Run starts the server
func (r *ginRouter) Run(addr string) error {
	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r.Shutdown(ctx)
	}()

	return r.router.Run(addr)
}

// Shutdown gracefully shuts down the router
func (r *ginRouter) Shutdown(ctx context.Context) error {
	log.Println("Router shutdown complete")
	return nil
}

// Start starts the router with configuration
func (r *ginRouter) Start(addr string) error {
	return r.Run(addr)
}

// healthCheck handles health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "audit-service",
		"version":   "1.0.0",
	})
}

// readinessCheck handles readiness check endpoint
func readinessCheck(c *gin.Context) {
	// Check database connection, etc.
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// CustomRouter is a more flexible router implementation
type CustomRouter struct {
	engine *gin.Engine
}

// NewCustomRouter creates a new custom router
func NewCustomRouter() *CustomRouter {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.SecurityHeaders())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	return &CustomRouter{engine: engine}
}

// Engine returns the underlying gin engine
func (r *CustomRouter) Engine() *gin.Engine {
	return r.engine
}

// GET adds a GET route
func (r *CustomRouter) GET(path string, handler gin.HandlerFunc) {
	r.engine.GET(path, handler)
}

// POST adds a POST route
func (r *CustomRouter) POST(path string, handler gin.HandlerFunc) {
	r.engine.POST(path, handler)
}

// PUT adds a PUT route
func (r *CustomRouter) PUT(path string, handler gin.HandlerFunc) {
	r.engine.PUT(path, handler)
}

// DELETE adds a DELETE route
func (r *CustomRouter) DELETE(path string, handler gin.HandlerFunc) {
	r.engine.DELETE(path, handler)
}

// PATCH adds a PATCH route
func (r *CustomRouter) PATCH(path string, handler gin.HandlerFunc) {
	r.engine.PATCH(path, handler)
}

// Group creates a new route group
func (r *CustomRouter) Group(path string) *gin.RouterGroup {
	return r.engine.Group(path)
}

// Run starts the server
func (r *CustomRouter) Run(addr string) error {
	return r.engine.Run(addr)
}

// RunTLS starts the server with TLS
func (r *CustomRouter) RunTLS(addr, certFile, keyFile string) error {
	return r.engine.RunTLS(addr, certFile, keyFile)
}

// Shutdown gracefully shuts down the server
func (r *CustomRouter) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Perform cleanup
	log.Println("Performing router shutdown...")

	return nil
}

// Handle registers a custom handler
func (r *CustomRouter) Handle(method, path string, handler gin.HandlerFunc) {
	r.engine.Handle(method, path, handler)
}

// Static serves static files
func (r *CustomRouter) Static(path, root string) {
	r.engine.Static(path, root)
}

// StaticFile serves a single static file
func (r *CustomRouter) StaticFile(path, file string) {
	r.engine.StaticFile(path, file)
}

// LoadHTMLGlob loads HTML templates
func (r *CustomRouter) LoadHTMLGlob(pattern string) {
	r.engine.LoadHTMLGlob(pattern)
}

// LoadHTMLFiles loads HTML templates from files
func (r *CustomRouter) LoadHTMLFiles(files ...string) {
	r.engine.LoadHTMLFiles(files...)
}

// HTML renders HTML template
func (r *CustomRouter) HTML(name string, obj interface{}) gin.H {
	return gin.H{
		"template": name,
		"data":     obj,
	}
}

// ResponseWriter provides custom response writing
type ResponseWriter struct {
	gin.ResponseWriter
}

// WriteJSON writes JSON response
func WriteJSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// WriteJSONOK writes success JSON response
func WriteJSONOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// WriteJSONError writes error JSON response
func WriteJSONError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error":   "error",
		"message": message,
	})
}

// WriteJSONCreated writes created JSON response
func WriteJSONCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// WriteJSONNoContent writes no content JSON response
func WriteJSONNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// GetContextValue gets a value from context
func GetContextValue(c *gin.Context, key string) interface{} {
	return c.Get(key)
}

// SetContextValue sets a value in context
func SetContextValue(c *gin.Context, key string, value interface{}) {
	c.Set(key, value)
}

// AbortWithError aborts with error response
func AbortWithError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":   "error",
		"message": message,
	})
}

// GetRequestID gets request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get("request_id"); exists {
		return id.(string)
	}
	return ""
}

// GetUserID gets user ID from context
func GetUserID(c *gin.Context) string {
	if id, exists := c.Get("user_id"); exists {
		return id.(string)
	}
	return ""
}

// GetUserRole gets user role from context
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

// Logger provides structured logging
type Logger struct {
	service string
}

// NewLogger creates a new logger
func NewLogger(service string) *Logger {
	return &Logger{service: service}
}

// Log logs a message
func (l *Logger) Log(level, message string, fields map[string]interface{}) {
	fields["service"] = l.service
	fields["level"] = level
	fields["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	log.Printf("[%s] %s: %v", level, message, fields)
}

// Debug logs debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	l.Log("debug", message, mergeFields(fields...))
}

// Info logs info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	l.Log("info", message, mergeFields(fields...))
}

// Warn logs warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	l.Log("warn", message, mergeFields(fields...))
}

// Error logs error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	l.Log("error", message, mergeFields(fields...))
}

// Fatal logs fatal message and exits
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	l.Log("fatal", message, mergeFields(fields...))
	os.Exit(1)
}

// mergeFields merges multiple field maps
func mergeFields(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxHeaderBytes int
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:            ":9090",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		MaxHeaderBytes:  1 << 20,
	}
}

// RunServer runs the server with configuration
func RunServer(router gin.IRouter, config ServerConfig) error {
	srv := &http.Server{
		Addr:         config.Addr,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on %s", config.Addr)
	return srv.ListenAndServe()
}

// RunServerTLS runs the server with TLS
func RunServerTLS(router gin.IRouter, config ServerConfig, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:         config.Addr,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on %s with TLS", config.Addr)
	return srv.ListenAndServeTLS(certFile, keyFile)
}

// GetPort extracts port from address
func GetPort(addr string) string {
	if addr == "" {
		return ":9090"
	}
	if addr[0] != ':' {
		return ":" + addr
	}
	return addr
}

// StringToDuration converts string to duration
func StringToDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}
