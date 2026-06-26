package main

import (
	"go-app/internal/initianlize"
)

func main() {
	// fmt.Println("Hello World")
	// r := routers.NewRouter() // gọi router
	// r.Run(":8080")           // listen and serve on 0.0.0.0:8080

	// Run all initianlize functions
	initianlize.Run() // run on port 8081
}
