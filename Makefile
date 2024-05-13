GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
NETSPEED_VERSION := ${NETSPEED_VERSION}

.PHONY: all build install

all: build install

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: build OS ARCH
build: guard-NETSPEED_VERSION mod-tidy clean
	@echo "================================================="
	@echo "Building netspeed"
	@echo "=================================================\n"

	@if [ ! -d "${GOOS}" ]; then \
		mkdir "${GOOS}"; \
	fi
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o "${GOOS}/netspeed"
	sleep 2
	tar -C "${GOOS}" -czvf "netspeed_${NETSPEED_VERSION}_${GOOS}_${GOARCH}.tgz" netspeed; \

.PHONY: clean
clean:
	@echo "================================================="
	@echo "Cleaning netspeed"
	@echo "=================================================\n"
	@for OS in darwin linux; do \
		if [ -f $${OS}/netspeed ]; then \
			rm -f $${OS}/netspeed; \
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
	@echo "Installing netspeed in ${GOPATH}/bin"
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