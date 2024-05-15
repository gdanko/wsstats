GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
WSSTATS_VERSION := ${WSSTATS_VERSION}

.PHONY: all build install

all: build install

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: build OS ARCH
build: guard-WSSTATS_VERSION mod-tidy clean
	@echo "================================================="
	@echo "Building wsstats"
	@echo "=================================================\n"

	@if [ ! -d "${GOOS}" ]; then \
		mkdir "${GOOS}"; \
	fi
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o "${GOOS}/wsstats"
	sleep 2
	tar -C "${GOOS}" -czvf "wsstats_${WSSTATS_VERSION}_${GOOS}_${GOARCH}.tgz" wsstats; \

.PHONY: clean
clean:
	@echo "================================================="
	@echo "Cleaning wsstats"
	@echo "=================================================\n"
	@for OS in darwin linux; do \
		if [ -f $${OS}/wsstats ]; then \
			rm -f $${OS}/wsstats; \
		fi; \
	done

.PHONY: clean-all
clean-all: clean
	@echo "================================================="
	@echo "Cleaning tarballs"
	@echo "=================================================\n"
	@rm -f *.tgz 2>/dev/null

.PHONY: install
install:
	@echo "================================================="
	@echo "Installing wsstats in ${GOPATH}/bin"
	@echo "=================================================\n"

	go install -race

#
# General targets
#
guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi