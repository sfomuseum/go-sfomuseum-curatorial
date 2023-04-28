GOMOD=vendor

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/lookup cmd/lookup/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/assign-exhibition-gallery cmd/assign-exhibition-gallery/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/supersede-exhibition cmd/supersede-exhibition/main.go

compile-all:
	@make compile-publicart-data
	@make compile-exhibitions-data
	@make compile-collection-data
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/lookup cmd/lookup/main.go

compile-publicart-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-publicart-data/main.go

compile-exhibitions-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-exhibitions-data/main.go

compile-collection-data:
	go run -mod $(GOMOD) -ldflags="-s -w" cmd/compile-collection-data/main.go
