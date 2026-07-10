package main

import (
	"go-app/internal/initianlize"
)

// @title Go App API
// @version 1.0
// @description This is a sample server API for Go App.
// @host localhost:8081
// @BasePath /v1/api

func main() {
	// fmt.Println("Hello World")
	// r := routers.NewRouter() // gọi router
	// r.Run(":8080")           // listen and serve on 0.0.0.0:8080

	// Run all initianlize functions
	initianlize.Run() // run on port 8081
}
