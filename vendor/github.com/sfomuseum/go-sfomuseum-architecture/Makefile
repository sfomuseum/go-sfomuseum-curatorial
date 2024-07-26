GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")
LDFLAGS=-s -w


cli:
	@make cli-lookup
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/supersede-gallery cmd/supersede-gallery/main.go

cli-lookup:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" -o bin/lookup cmd/lookup/main.go

cli-complex:
	go build -mod $(GOMOD) -ldflags="$(LDFLAGS)" --tags json1 -o bin/current-complex cmd/current-complex/main.go

compile:
	@make compile-gates
	@make compile-galleries
	@make compile-terminals
	@make cli-lookup

compile-gates:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-gates-data/main.go

compile-terminals:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-terminals-data/main.go

compile-galleries:
	go run -mod $(GOMOD) -ldflags="$(LDFLAGS)" cmd/compile-galleries-data/main.go
