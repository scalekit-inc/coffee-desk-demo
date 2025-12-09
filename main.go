package main

import (
	"coffee-desk-demo/config"
	"coffee-desk-demo/database"
	"coffee-desk-demo/handlers"
	"coffee-desk-demo/internal/templates"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	// Initialize the global ScaleKit client
	if err := handlers.InitializeScaleKitClient(); err != nil {
		panic(fmt.Sprintf("Failed to initialize ScaleKit client: %v", err))
	}

	// Initialize connections handler
	connectionsHandler, err := handlers.NewConnectionsHandler()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize connections handler: %v", err))
	}

	// Initialize connected accounts handler
	connectedAccountsHandler, err := handlers.NewConnectedAccountsHandler()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize connected accounts handler: %v", err))
	}

	// Initialize database
	if err := database.InitDatabase(); err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	// Run database migrations
	if err := database.AutoMigrate(); err != nil {
		panic(fmt.Sprintf("Failed to run database migrations: %v", err))
	}

	port := config.GetEnv("PORT", "8080")

	r := gin.Default()

	// Configure CORS (can be reduced since we're serving from same origin)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins in production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Enforce HTTPS
	r.Use(func(c *gin.Context) {
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	})

	// Set up HTML renderer with templates
	r.SetHTMLTemplate(templates.Templates)

	// Serve static assets (CSS, JS, images from React build)
	assets, err := templates.Assets()
	if err != nil {
		panic(err)
	}
	r.StaticFS("/assets", http.FS(assets))

	// Serve uploads (logos, images, etc.)
	uploads, err := templates.Uploads()
	if err != nil {
		panic(err)
	}
	r.StaticFS("/uploads", http.FS(uploads))

	// Serve specific root files
	rootFiles, err := templates.RootFiles()
	if err != nil {
		panic(err)
	}

	// Serve favicon.ico (using SVG format - modern browsers support it)
	r.GET("/favicon.ico", func(c *gin.Context) {
		// Serve the coffee desk favicon SVG from uploads
		file, err := uploads.Open("coffee-desk-favicon.svg")
		if err != nil {
			c.Status(404)
			return
		}
		defer file.Close()
		c.DataFromReader(200, -1, "image/svg+xml", file, nil)
	})

	// Serve robots.txt
	r.GET("/robots.txt", func(c *gin.Context) {
		file, err := rootFiles.Open("robots.txt")
		if err != nil {
			c.Status(404)
			return
		}
		defer file.Close()
		c.DataFromReader(200, -1, "text/plain", file, nil)
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})
		api.GET("/authorize", handlers.AuthorizeHandler)
		api.GET("/scalekit/callback", handlers.CallbackHandler)
		api.GET("/session", handlers.SessionHandler)
		api.GET("/logout", handlers.LogoutHandler)
		api.POST("/webhooks/scalekit", handlers.ScalekitWebhookHandler)
		api.GET("/workspace/members", handlers.GetWorkspaceMembersHandler)
		api.POST("/workspace/members", handlers.CreateWorkspaceMemberHandler)
		api.DELETE("/workspace/members/:member_id", handlers.DeleteWorkspaceMemberHandler)
		api.POST("/workspace/members/resend-invite", handlers.ResendInviteHandler)
		api.PUT("/workspace", handlers.UpdateWorkspaceHandler)
		api.POST("/workspace/create", handlers.CreateWorkspaceHandler)
		api.GET("/workspace/domains", handlers.GetDomainsHandler)
		api.POST("/workspace/domains", handlers.CreateDomainsHandler)
		api.DELETE("/workspace/domains/:domain_id", handlers.DeleteDomainHandler)
		api.PUT("/user/profile", handlers.UpdateUserProfileHandler)
		api.GET("/portal/link", handlers.GetPortalLinkHandler)
		api.GET("/scalekit/environment-url", handlers.GetScaleKitEnvironmentURLHandler)
		api.GET("/scalekit/passkeys", handlers.RedirectToPasskeysHandler)
		api.POST("/workspace/onboarding", handlers.OnboardingHandler)

		//
		// Connections routes - registers /api/connections/* endpoints
		connectionsHandler.RegisterRoutes(api)

		// Connected accounts routes - registers GET /api/connections?connector=slack|github
		connectedAccountsHandler.RegisterRoutes(api)

		// Project Management API routes
		api.GET("/projects", handlers.GetProjectsHandler)
		api.POST("/projects", handlers.CreateProjectHandler)
		api.GET("/projects/:id", handlers.GetProjectHandler)
		api.PUT("/projects/:id", handlers.UpdateProjectHandler)
		api.DELETE("/projects/:id", handlers.DeleteProjectHandler)
		api.GET("/tasks", handlers.GetTasksHandler)
		api.POST("/tasks", handlers.CreateTaskHandler)
		api.GET("/tasks/:id", handlers.GetTaskHandler)
		api.PUT("/tasks/:id", handlers.UpdateTaskHandler)
		api.DELETE("/tasks/:id", handlers.DeleteTaskHandler)
	}

	// Serve React app for all other routes
	r.NoRoute(func(c *gin.Context) {
		// You can inject server-side data here if needed
		data := gin.H{
			"Title": "Biz Beacon",
			// Add any server-side data you want to pass to React
		}
		c.HTML(200, "main", data)
	})

	r.Run(":" + port)
}
