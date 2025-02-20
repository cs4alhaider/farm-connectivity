package main

import (
	"farm-connectivity/database"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func pingGoogle() (float64, bool) {
	start := time.Now()
	resp, err := http.Get("https://google.com")
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()

	latency := time.Since(start).Seconds() * 1000 // Convert to milliseconds
	return latency, resp.StatusCode == 200
}

func startPingMonitor(app *fiber.App) {
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		for range ticker.C {
			latency, success := pingGoogle()
			err := database.SavePingResult(latency, success)
			if err != nil {
				log.Printf("Error saving ping result: %v", err)
			}
		}
	}()
}

func main() {
	// Initialize database
	database.InitDB()

	// Initialize template engine
	engine := html.New("./views", ".html")

	// Create fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Serve static files
	app.Static("/static", "./static")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Get("/api/ping-results", func(c *fiber.Ctx) error {
		results, err := database.GetPingResults(100) // Get last 100 results
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch ping results"})
		}
		return c.JSON(results)
	})

	// Start ping monitor
	startPingMonitor(app)

	// Start server
	log.Fatal(app.Listen(":3000"))
}
