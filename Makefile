cli:
	go build -mod vendor -o bin/lookup cmd/lookup/main.go
	go build -mod vendor -o bin/assign-exhibition-gallery cmd/assign-exhibition-gallery/main.go

compile-all:
	@make compile-publicart-data
	@make compile-exhibitions-data
	@make compile-collection-data
	go build -mod vendor -o bin/lookup cmd/lookup/main.go

compile-publicart-data:
	go run -mod vendor cmd/compile-publicart-data/main.go

compile-exhibitions-data:
	go run -mod vendor cmd/compile-exhibitions-data/main.go

compile-collection-data:
	go run -mod vendor cmd/compile-collection-data/main.go
