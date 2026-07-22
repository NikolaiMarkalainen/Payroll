.PHONY: run build test clean air

BINARY := bin/payroll

build:
	@mkdir -p bin
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) ./cmd/

run: build
	@echo "Starting Palkkatarkistus..."
	./$(BINARY)

test:
	go test ./...

# Live reload (requires: go install github.com/air-verse/air@latest)
air:
	@command -v air >/dev/null 2>&1 || { \
		echo "air ei ole PATH:ssa. Asenna: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	air

clean:
	rm -rf bin tmp
