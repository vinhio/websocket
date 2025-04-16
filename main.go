package main

import (
	"github.com/gflydev/core"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	app := core.New()

	// Register router
	app.RegisterRouter(func(g core.IFly) {
		g.GET("/ws", NewWSHandler())
	})

	app.Run()
}
