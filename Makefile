NAME=blast
BUILD_DIR ?= bin
BUILD_FLAGS=-ldflags="-s -w"
BUILD_SRC=.

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: all clean test build tools format pre-commit
all: clean deps test build

deps: tools
	@printf "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)\n"
	@go mod vendor

build:
	@echo "$(OK_COLOR)==> Building the application...$(NO_COLOR)"
	@CGO_ENABLED=0 go build -v $(BUILD_FLAGS) -o "$(BUILD_DIR)/$(NAME)" "$(BUILD_SRC)"

clean:
	@rm -rf ./bin

test: test-unit

test-unit:
	@echo "$(OK_COLOR)==> Running the unit tests$(NO_COLOR)"
	@go test -race -tags unit -cover ./...

format: tools
	@echo "$(OK_COLOR)>> [$@] go vet: running$(NO_COLOR)"
	@go vet ./...

	@echo "$(OK_COLOR)>> [$@] gci: running$(NO_COLOR)"
	@gci -w pkg

	@echo "$(OK_COLOR)>> [$@] gofumpt: running$(NO_COLOR)"
	@gofumpt -w pkg

tools:
	@if ! command -v gci > /dev/null ; then \
		echo ">> [$@]: gci not found: installing"; \
		go install github.com/daixiang0/gci@latest; \
	fi

	@if ! command -v gofumpt > /dev/null ; then \
		echo ">> [$@]: gofumpt not found: installing"; \
		go install mvdan.cc/gofumpt@latest; \
	fi
