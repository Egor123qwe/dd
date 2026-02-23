# Сборка мерчант-клиента и runtimedaemon для macOS и Windows.
# Результат: merchant_build/merchant_macos/ или merchant_build/merchant_windows/
# Запуск: из папки client/ — make build-mac или make build-windows

BUILD_DIR     = merchant_build
MACOS_DIR     = $(BUILD_DIR)/merchant_macos
WINDOWS_DIR   = $(BUILD_DIR)/merchant_windows
RD_SUBDIR    = runtimedaemon

# Использовать локальный Go, не качать тулчейн (нужно при GOSUMDB=off)
GO_ENV      = GOTOOLCHAIN=local
# macOS: darwin, по умолчанию arm64 (Apple Silicon)
GO_MACOS    = $(GO_ENV) GOOS=darwin GOARCH=arm64
# Windows: amd64
GO_WINDOWS  = $(GO_ENV) GOOS=windows GOARCH=amd64

.PHONY: build-mac build-windows clean help

help:
	@echo "Сборка мерчант-клиента (merchant_build/):"
	@echo "  make build-mac     — merchant_build/merchant_macos/ (merchantclient + runtimedaemon/, darwin arm64)"
	@echo "  make build-windows — merchant_build/merchant_windows/ (merchantclient.exe + runtimedaemon/, windows amd64)"
	@echo "  make clean         — удалить merchant_build/"

build-mac: $(MACOS_DIR)/merchantclient $(MACOS_DIR)/$(RD_SUBDIR)/runtimedaemon
	@test -f $(MACOS_DIR)/.env || echo 'BACKEND_HOST=' > $(MACOS_DIR)/.env
	@echo "Готово: $(MACOS_DIR)/"

build-win: $(WINDOWS_DIR)/merchantclient.exe $(WINDOWS_DIR)/$(RD_SUBDIR)/runtimedaemon.exe
	@test -f $(WINDOWS_DIR)/.env || echo 'BACKEND_HOST=' > $(WINDOWS_DIR)/.env
	@echo "Готово: $(WINDOWS_DIR)/"

$(MACOS_DIR):
	@mkdir -p $(MACOS_DIR)/$(RD_SUBDIR)

$(WINDOWS_DIR):
	@mkdir -p $(WINDOWS_DIR)/$(RD_SUBDIR)

$(MACOS_DIR)/merchantclient: $(MACOS_DIR)
	cd merchantclient && $(GO_MACOS) go build -o ../$(MACOS_DIR)/merchantclient cmd/merchantclient/main.go

$(MACOS_DIR)/$(RD_SUBDIR)/runtimedaemon: $(MACOS_DIR)
	cd runtimedaemonv2 && $(GO_MACOS) go build -o ../$(MACOS_DIR)/$(RD_SUBDIR)/runtimedaemon cmd/runtimedaemon/main.go

$(WINDOWS_DIR)/merchantclient.exe: $(WINDOWS_DIR)
	cd merchantclient && $(GO_WINDOWS) go build -o ../$(WINDOWS_DIR)/merchantclient.exe cmd/merchantclient/main.go

$(WINDOWS_DIR)/$(RD_SUBDIR)/runtimedaemon.exe: $(WINDOWS_DIR)
	cd runtimedaemonv2 && $(GO_WINDOWS) go build -o ../$(WINDOWS_DIR)/$(RD_SUBDIR)/runtimedaemon.exe cmd/runtimedaemon/main.go

clean:
	rm -rf $(BUILD_DIR)
	@echo "Папка $(BUILD_DIR) удалена."
