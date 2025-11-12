# Имя бинарного файла
BINARY_NAME=server

# Путь к папке с исходным кодом
SRC_DIR=.

# Флаги для сборки (с оптимизацией и без отладочной информации)
GO_BUILD_FLAGS=-ldflags "-s -w -extldflags '-static'"

# Цель для форматирования кода (gofmt)
fmt:
	@gofmt -w $(SRC_DIR)

# Цель для исправления импортов (goimports)
imports:
	@goimports -w $(SRC_DIR)

# Цель для сборки текущей ОС
build: fmt imports
	@CGO_ENABLED=0 go build -a $(GO_BUILD_FLAGS) -o $(BINARY_NAME) $(SRC_DIR)

# Цель для сборки для Windows 7 32-bit
build_windows: fmt imports
	@GOOS=windows GOARCH=386 go build $(GO_BUILD_FLAGS) -o $(BINARY_NAME)-win32.exe $(SRC_DIR)

# Цель для запуска
run: build
	@./$(BINARY_NAME)

# Цель для чистки
clean:
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-win32.exe

# Хелп: вывод справки при отсутствии указания цели
help:
	@echo "Доступные цели:"
	@echo "  build          - Сборка проекта для текущей ОС"
	@echo "  build_windows  - Сборка проекта для Windows 7 32-bit"
	@echo "  run            - Запуск собранного приложения"
	@echo "  clean          - Удаление бинарных файлов"
	@echo "  fmt            - Форматирование кода с помощью gofmt"
	@echo "  imports        - Автоматическое исправление импортов с помощью goimports"

# По умолчанию выводится справка, если цель не указана
.DEFAULT_GOAL := help

.PHONY: build run clean build_windows fmt imports help
