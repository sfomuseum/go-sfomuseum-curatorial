compile-all:
	@make compile-publicart-data
	@make compile-exhibitions-data
	go build -mod vendor -o bin/lookup cmd/lookup/main.go

compile-publicart-data:
	go run -mod vendor cmd/compile-publicart-data/main.go

compile-exhibitions-data:
	go run -mod vendor cmd/compile-exhibitions-data/main.go
