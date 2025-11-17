### -----------------------
# --- Make variables
### -----------------------

# only evaluated if required by a recipe
# http://make.mad-scientist.net/deferred-simple-variable-expansion/

# go module name (as in go.mod)
GO_MODULE_NAME = $(eval GO_MODULE_NAME := $$(shell \
	(mkdir -p tmp data 2> /dev/null && cat .modulename 2> /dev/null) \
	|| (gsdev modulename 2> /dev/null | tee .modulename) || echo "unknown" \
))$(GO_MODULE_NAME)

# Common infos
MODULE_NAME := $(eval MODULE_NAME := $(shell echo $(GO_MODULE_NAME) | awk -F'/' '{print $$3}'))$(MODULE_NAME)
MODULE_VERSION := $(shell git describe --tags --always)

# https://medium.com/the-go-journey/adding-version-information-to-go-binaries-e1b79878f6f2
ARG_COMMIT = $(eval ARG_COMMIT := $$(shell \
	(git rev-list -1 HEAD 2> /dev/null) \
	|| (echo "unknown") \
))$(ARG_COMMIT)

ARG_BUILD_DATE = $(eval ARG_BUILD_DATE := $$(shell \
	(date -Is 2> /dev/null || date 2> /dev/null || echo "unknown") \
))$(ARG_BUILD_DATE)

# https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
LDFLAGS = $(eval LDFLAGS := "\
-X '$(GO_MODULE_NAME)/internal/config.ModuleName=$(MODULE_NAME)'\
-X '$(GO_MODULE_NAME)/internal/config.Commit=$(ARG_COMMIT)'\
-X '$(GO_MODULE_NAME)/internal/config.BuildDate=$(ARG_BUILD_DATE)'\
")$(LDFLAGS)

# Debian infos
DEB_REGISTRY ?= http://172.31.1.1:8888/repository/hosted/
DEB_USER ?= c
DEB_PASSWORD ?= 1
DEB_PACKAGE ?= ${MODULE_NAME}
DEB_VERSION ?= ${MODULE_VERSION}
DEB_ARCHITECTURE ?= amd64
DEB_MAINTAINER ?= Huy Le <test@viettel.com.vn>
DEB_PRE_DEPENDS ?=
DEB_DEPENDS ?=
DEB_RECOMMENDS ?= mongodb-org, nats-server
DEB_DESCRIPTION ?= Service for managing ${DEB_PACKAGE}

# Docker infos
DOCKER_REGISTRY ?= harbor.vht.vn
DOCKER_PROJECT ?= c4i
DOCKER_USER ?= admin
DOCKER_PASSWORD ?= 1
DOCKER_IMAGE_NAME ?= ${MODULE_NAME}
DOCKER_IMAGE_VERSION ?= ${MODULE_VERSION}
DOCKER_IMAGE := ${DOCKER_REGISTRY}/${DOCKER_PROJECT}/${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_VERSION}

# Service infos
DATA_DIR ?= data
STYLE_TIMEOUT ?= 1080000
MAX_STYLE_CONCURRENCY ?= 16

# Golang build infos
CGO_ENABLED ?= 0

### -----------------------
# --- End variables
### -----------------------

### -----------------------
# --- Start dependencies
### -----------------------

proto:
	@echo "running proto..."

	@rm -rf pkg/pb/*.go

	@protoc --proto_path=/usr/local/include --proto_path=. --go_out=. --go-grpc_out=. pkg/proto/*.proto

	@cd pkg && ./custom.generate.sh

tidy:
	@echo "running tidy..."

	@go mod tidy

vendor:
	@echo "running vendor..."

	@go mod vendor -o vendor-tmp

	@cp -rvf vendor-tmp/* vendor/
	@rm -rf vendor-tmp

swag:
	@echo "creating swagger..."

	@swag init --parseDependency --parseInternal --parseDepth 100 -g main.go
	@swag fmt

### -----------------------
# --- End dependencies
### -----------------------

### -----------------------
# --- Start buildings
### -----------------------

go-build:
	@echo "building binary..."
	@CGO_ENABLED=${CGO_ENABLED} go build -mod=vendor -ldflags ${LDFLAGS} -o bin/${MODULE_NAME}

build:
	@echo "building..."

	@make proto
	@make tidy
	@make vendor
	@make go-build

### -----------------------
# --- End buildings
### -----------------------

### -----------------------
# --- Helpers
### -----------------------

clean: ##- Cleans ./tmp and ./api/tmp folder.
	@echo "cleaning..."

	@rm -rf tmp 2> /dev/null
	@rm -rf api/tmp 2> /dev/null

get-module-name: ##- Prints current go module-name (pipeable).
	@echo "${GO_MODULE_NAME}"

set-module-name: ##- Wizard to set a new go module-name.
	@echo "setting module name..."

	@rm -rf .modulename
	@echo "Enter new go module-name:" \
		&& read new_module_name \
		&& echo "new go module-name: '$${new_module_name}'" \
		&& echo -n "Are you sure? [y/N]" \
		&& read ans && [ $${ans:-N} = y ] \
		&& echo -n "Please wait..." \
		&& find . -not -path '*/\.*' -not -path './Makefile' -type f -exec sed -i "s|${GO_MODULE_NAME}|192.168.0.1/name-service/$${new_module_name}|g" {} \; \
		&& sed -i "s|/${MODULE_NAME}|/$${new_module_name}|g" './Dockerfile' \
		&& echo "172.21.5.249/map-server/$${new_module_name}" >> .modulename \
		&& echo "new go module-name: '$${new_module_name}'!"

get-go-ldflags: ##- (opt) Prints used -ldflags as evaluated in Makefile used in make go-build
	@echo ${LDFLAGS}

tools: ##- (opt) Install packages as specified in tools.go.
	@echo "installing tools..."

	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -P $${nproc} -L 1 -tI % go install %

### -----------------------
# --- Start dockers
### -----------------------

login-docker-registry:
	@echo 'logining docker to registry "${DOCKER_REGISTRY}"...'

	docker login \
		${DOCKER_REGISTRY} \
		-u ${DOCKER_USER} \
		-p ${DOCKER_PASSWORD}

build-docker-image:
	@echo 'building docker image "${DOCKER_IMAGE}"...'

	docker build \
		-t ${DOCKER_IMAGE} \
		-f ./Dockerfile .

push-docker-image:
	@echo 'pushing docker image "${DOCKER_IMAGE}"...'

	docker push \
		${DOCKER_IMAGE}

clean-docker-image:
	@echo 'cleaning docker image "${DOCKER_IMAGE}"...'

	docker rmi -f \
		${DOCKER_IMAGE}

release-docker-image:
	@echo 'releasing docker image "${DOCKER_IMAGE}"...'

	@make build-docker-image
	@make push-docker-image
	@make clean-docker-image

### -----------------------
# --- End dockers
### -----------------------

### -----------------------
# --- Start debians
### -----------------------

build-debian:
	@echo "building debian package..."

	@make go-build

	$(eval DEB_ROOT_FOLDER = deb/$(DEB_PACKAGE)-$(DEB_ARCHITECTURE)-$(DEB_VERSION))
	$(eval DEB_FOLDER = $(DEB_ROOT_FOLDER)/DEBIAN)
	$(eval DEB_BIN_FOLDER = $(DEB_ROOT_FOLDER)/usr/local/bin)
	$(eval DEB_CONFIG_FOLDER = $(DEB_ROOT_FOLDER)/usr/local/etc/$(DEB_PACKAGE))
	$(eval DEB_SERVICE_CONFIG_FOLDER = $(DEB_ROOT_FOLDER)/usr/lib/systemd/system)

	$(eval DEB_SERVICE_FILE = $(DEB_SERVICE_CONFIG_FOLDER)/$(DEB_PACKAGE).service)
	$(eval DEB_SERVICE_ENV_FILE = $(DEB_CONFIG_FOLDER)/env)
	$(eval DEB_CONTROL_FILE = $(DEB_FOLDER)/control)
	$(eval DEB_POSTINST_FILE = $(DEB_FOLDER)/postinst)
	$(eval DEB_PRERM_FILE = $(DEB_FOLDER)/prerm)

	@mkdir -p \
		${DEB_ROOT_FOLDER} \
		${DEB_FOLDER} \
		${DEB_BIN_FOLDER} \
		${DEB_CONFIG_FOLDER} \
		${DEB_SERVICE_CONFIG_FOLDER}

	@cp deb_template/name-arch-version/usr/lib/systemd/system/name.service ${DEB_SERVICE_FILE}
	@cp -r deb_template/name-arch-version/DEBIAN/* ${DEB_FOLDER}

	@cp -r etc/* ${DEB_CONFIG_FOLDER}/
	@cp -r bin/* ${DEB_BIN_FOLDER}/
	@cp -r web ${DEB_BIN_FOLDER}/

	@echo 'DATA_DIR=${DATA_DIR}' >> ${DEB_SERVICE_ENV_FILE}
	@echo 'STYLE_CONCURRENCY=${STYLE_CONCURRENCY}' >> ${DEB_SERVICE_ENV_FILE}
	@echo 'MAX_STYLE_CONCURRENCY=${MAX_STYLE_CONCURRENCY}' >> ${DEB_SERVICE_ENV_FILE}

	@sed -i 's#SERVICE_DESCRIPTION#${DEB_DESCRIPTION}#g' ${DEB_SERVICE_FILE}
	@sed -i 's#SERVICE_BIN#/usr/local/bin/${DEB_PACKAGE}#g' ${DEB_SERVICE_FILE}
	@sed -i 's#SERVICE_CONFIG#/usr/local/etc/${DEB_PACKAGE}#g' ${DEB_SERVICE_FILE}
	@sed -i 's#SERVICE_ENV#/usr/local/etc/${DEB_PACKAGE}/env#g' ${DEB_SERVICE_FILE}

	@sed -i 's#DEB_PACKAGE#${DEB_PACKAGE}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_VERSION#${DEB_VERSION}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_ARCHITECTURE#${DEB_ARCHITECTURE}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_MAINTAINER#${DEB_MAINTAINER}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_PRE_DEPENDS#${DEB_PRE_DEPENDS}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_DEPENDS#${DEB_DEPENDS}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_RECOMMENDS#${DEB_RECOMMENDS}#g' ${DEB_CONTROL_FILE}
	@sed -i 's#DEB_DESCRIPTION#${DEB_DESCRIPTION}#g' ${DEB_CONTROL_FILE}

	@sed -i 's#DEB_PACKAGE#${DEB_PACKAGE}#g' ${DEB_POSTINST_FILE}
	@sed -i 's#DEB_PACKAGE#${DEB_PACKAGE}#g' ${DEB_PRERM_FILE}

	@chmod -R 775 ${DEB_ROOT_FOLDER}

	@dpkg-deb --build --root-owner-group $(DEB_ROOT_FOLDER)

push-debian:
	$(eval DEB_FILE = deb/$(DEB_PACKAGE)-$(DEB_ARCHITECTURE)-$(DEB_VERSION).deb)

	@echo 'pushing debian package "${DEB_FILE}"...'

	@curl -u "${DEB_USER}:${DEB_PASSWORD}" -H "Content-Type: multipart/form-data" --data-binary "@./${DEB_FILE}" "${DEB_REGISTRY}"

release-debian:
	$(eval DEB_FILE = deb/$(DEB_PACKAGE)-$(DEB_ARCHITECTURE)-$(DEB_VERSION).deb)

	@echo 'releasing debian "${DEB_FILE}"...'

	@make build-debian
	@make push-debian

### -----------------------
# --- End debians
### -----------------------

### -----------------------
# --- Start devs
### -----------------------

server:
	@echo "starting server..."

	@make swag
	@make go-build
	@./bin/${MODULE_NAME} start ./etc-dev

### -----------------------
# --- End devs
### -----------------------

.PHONY: tools vendor proto
