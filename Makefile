.PHONY: run run-demo build test clean air

BINARY := bin/payroll

build:
	@mkdir -p bin
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) ./cmd/

run: build
	@echo "Starting Palkkatarkistus..."
	./$(BINARY)

# Demo: roster Jul–Aug 2026 preloaded, opens Vuorot tab
run-demo: build
	@echo "Starting Palkkatarkistus with demo shifts..."
	./$(BINARY) -demo

test:
	go test ./...

# Live reload with demo roster (requires: go install github.com/air-verse/air@latest).
# Does not force Vartiointi TES — Asetukset start on Oma (tyhjä).
air:
	@command -v air >/dev/null 2>&1 || { \
		echo "air ei ole PATH:ssa. Asenna: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	@echo "Starting air with demo shifts (-demo), TES = Oma..."
	air

clean:
	rm -rf bin tmp
