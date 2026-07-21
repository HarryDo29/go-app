package main

import (
	initianlize "go-app/internal/initialize"
)

// @title Go App API
// @version 1.0
// @description This is a sample server API for Go App.
// @host api.chat-app.website
// @BasePath /v1/api

func main() {
	// Run all initianlize functions
	initianlize.Run()
}
