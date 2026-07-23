.PHONY: run run-demo build test clean air \
	package package-linux package-darwin package-windows \
	install-linux uninstall-linux dist-binary \
	cross cross-linux cross-darwin cross-windows cross-all \
	fyne-tools

BINARY := bin/payroll
DIST := dist
APP_NAME := Palkkatarkistus
APP_ID := fi.palkkatarkistus.app
SRC := ./cmd/palkkatarkistus
PREFIX ?= $(HOME)/.local

# Native binary (current OS). Fast for local use; no desktop package metadata.
build:
	@mkdir -p bin
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) $(SRC)

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

# --- Packaging (desktop shortcut / .app / .exe) ---
# Needs: go install fyne.io/fyne/v2/cmd/fyne@latest
# Packages for the HOST OS only (use `make cross-*` for other platforms).
# Build the binary first so the executable is named Palkkatarkistus (not "cmd").

fyne-tools:
	@command -v fyne >/dev/null 2>&1 || { \
		echo "Asennetaan fyne CLI..."; \
		go install fyne.io/fyne/v2/cmd/fyne@latest; \
	}
	@command -v fyne-cross >/dev/null 2>&1 || { \
		echo "Asennetaan fyne-cross..."; \
		go install github.com/fyne-io/fyne-cross@latest; \
	}

# Always rebuild so install/package never ships a stale binary.
dist-binary:
	@mkdir -p $(DIST)
	go build -o $(DIST)/$(APP_NAME) $(SRC)

package: package-linux

package-linux: fyne-tools dist-binary
	@mkdir -p $(DIST)
	fyne package --os linux --name "$(APP_NAME)" --icon "$(CURDIR)/Icon.png" \
		--app-id fi.palkkatarkistus.app --executable $(DIST)/$(APP_NAME)
	@mv -f "$(APP_NAME).tar.xz" $(DIST)/ 2>/dev/null || true
	@echo "Linux package -> $(DIST)/$(APP_NAME).tar.xz"

# Install into ~/.local so "Palkkatarkistus" appears in the app menu search.
install-linux: package-linux
	@set -e; \
	tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	tar -xf $(DIST)/$(APP_NAME).tar.xz -C "$$tmpdir"; \
	root="$$tmpdir/$(APP_NAME)/usr/local"; \
	mkdir -p "$(PREFIX)/bin" "$(PREFIX)/share/applications" "$(PREFIX)/share/icons/hicolor/256x256/apps" "$(PREFIX)/share/pixmaps"; \
	install -m 755 "$$root/bin/$(APP_NAME)" "$(PREFIX)/bin/$(APP_NAME)"; \
	if [ -f "$$root/share/pixmaps/$(APP_ID).png" ]; then \
		install -m 644 "$$root/share/pixmaps/$(APP_ID).png" "$(PREFIX)/share/pixmaps/$(APP_ID).png"; \
		install -m 644 "$$root/share/pixmaps/$(APP_ID).png" "$(PREFIX)/share/icons/hicolor/256x256/apps/$(APP_ID).png"; \
	fi; \
	printf '%s\n' \
		'[Desktop Entry]' \
		'Type=Application' \
		'Name=$(APP_NAME)' \
		'Comment=TES-pohjainen palkkatarkistus' \
		'Exec=$(PREFIX)/bin/$(APP_NAME)' \
		'Icon=$(APP_ID)' \
		'Terminal=false' \
		'Categories=Office;Finance;' \
		'StartupWMClass=$(APP_NAME)' \
		'Keywords=palkka;TES;vuoro;payroll;' \
		> "$(PREFIX)/share/applications/$(APP_ID).desktop"; \
	chmod 644 "$(PREFIX)/share/applications/$(APP_ID).desktop"; \
	if command -v update-desktop-database >/dev/null 2>&1; then \
		update-desktop-database "$(PREFIX)/share/applications" >/dev/null 2>&1 || true; \
	fi; \
	if command -v gtk-update-icon-cache >/dev/null 2>&1; then \
		gtk-update-icon-cache -f -t "$(PREFIX)/share/icons/hicolor" >/dev/null 2>&1 || true; \
	fi; \
	echo "Asennettu: $(PREFIX)/bin/$(APP_NAME)"; \
	echo "Valikko:   $(PREFIX)/share/applications/$(APP_ID).desktop"; \
	echo "Hae sovellusvalikosta: $(APP_NAME)"

uninstall-linux:
	@rm -f "$(PREFIX)/bin/$(APP_NAME)" \
		"$(PREFIX)/share/applications/$(APP_ID).desktop" \
		"$(PREFIX)/share/pixmaps/$(APP_ID).png" \
		"$(PREFIX)/share/icons/hicolor/256x256/apps/$(APP_ID).png"
	@if command -v update-desktop-database >/dev/null 2>&1; then \
		update-desktop-database "$(PREFIX)/share/applications" >/dev/null 2>&1 || true; \
	fi
	@echo "Poistettu ~/.local -asennus ($(APP_NAME))."

package-darwin: fyne-tools dist-binary
	@mkdir -p $(DIST)
	fyne package --os darwin --name "$(APP_NAME)" --icon "$(CURDIR)/Icon.png" \
		--app-id fi.palkkatarkistus.app --executable $(DIST)/$(APP_NAME)
	@rm -rf $(DIST)/$(APP_NAME).app
	@mv -f "$(APP_NAME).app" $(DIST)/ 2>/dev/null || true
	@echo "macOS .app -> $(DIST)/$(APP_NAME).app (requires macOS host for full packaging)."

package-windows: fyne-tools
	@mkdir -p $(DIST)
	@echo "Windows native package: use make cross-windows from Linux, or build on Windows."
	fyne package --os windows --name "$(APP_NAME)" --icon "$(CURDIR)/Icon.png" \
		--app-id fi.palkkatarkistus.app --source-dir $(SRC)
	@mv -f "$(APP_NAME).exe" $(DIST)/ 2>/dev/null || true
	@echo "Windows .exe -> $(DIST)/"

# --- Cross builds via Docker (Linux host -> Linux / Windows / macOS) ---
# Needs: Docker + fyne-cross. macOS SDK optional for darwin (see fyne-cross docs).

cross: cross-all

cross-linux: fyne-tools
	@mkdir -p $(DIST)
	fyne-cross linux -arch=amd64,arm64 -name="$(APP_NAME)" -icon="$(CURDIR)/Icon.png" -app-id=fi.palkkatarkistus.app $(SRC)
	@cp -a fyne-cross/dist/linux-*/* $(DIST)/ 2>/dev/null || true
	@echo "Linux cross packages under fyne-cross/dist/ (copied to $(DIST)/ when present)."

cross-windows: fyne-tools
	@mkdir -p $(DIST)
	fyne-cross windows -arch=amd64 -name="$(APP_NAME)" -icon="$(CURDIR)/Icon.png" -app-id=fi.palkkatarkistus.app $(SRC)
	@cp -a fyne-cross/dist/windows-*/* $(DIST)/ 2>/dev/null || true
	@echo "Windows packages under fyne-cross/dist/."

cross-darwin: fyne-tools
	@mkdir -p $(DIST)
	fyne-cross darwin -arch=amd64,arm64 -name="$(APP_NAME)" -icon="$(CURDIR)/Icon.png" -app-id=fi.palkkatarkistus.app $(SRC)
	@cp -a fyne-cross/dist/darwin-*/* $(DIST)/ 2>/dev/null || true
	@echo "macOS packages under fyne-cross/dist/ (SDK may be required)."

cross-all: cross-linux cross-windows cross-darwin
	@echo "All cross builds done. See fyne-cross/dist/ and $(DIST)/."

clean:
	rm -rf bin tmp $(DIST) fyne-cross
