.PHONY: run build clean

BINARY := bin/payroll

build:
	@mkdir -p bin
	@echo "Building $(BINARY) (first Fyne/CGO build can take a few minutes)..."
	go build -o $(BINARY) ./cmd/

run: build
	@echo "Starting Payroll..."
	./$(BINARY)

clean:
	rm -rf bin
