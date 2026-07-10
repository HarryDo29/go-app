# app name
APP_NAME := server

# run dev
run:
	go run cmd/${APP_NAME}/main.go

# seed fake mongodb data
seed:
	go run ./cmd/seed
