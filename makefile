CC=go
RM=rm
MV=mv


SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
GOOS=linux
GOARCH=amd64

VERSION:=1.0
PREVIOUS_VERSION=$(shell echo $$((${VERSION} - 1)))

EXEC=bebop-ui-control

BUILD_TIME=`date +%FT%T%z`
PACKAGES :=

LIBS=

LDFLAGS=-ldflags "-w -race"

.DEFAULT_GOAL:=test

package: ${EXEC}
		@rm -rf ./deploy/${EXEC}_${VERSION}
		@cd deploy; mkdir ${EXEC}_${VERSION}
		@cp ${EXEC}config.json ./deploy/${EXEC}_${VERSION}/.
		@cp ./${EXEC}-${VERSION} ./deploy/${EXEC}_${VERSION}/.

		@cd deploy/${EXEC}_${VERSION}; tar -cvzf ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.tar.gz ${EXEC}-${VERSION}
		@mv ./deploy/${EXEC}_${VERSION}/${EXEC}-${GOOS}-${GOARCH}-${VERSION}.tar.gz deploy/.
		@echo "    Archive ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.tar.gz created"
		@rm -rf ./deploy/${EXEC}_${VERSION}

test: $(EXEC)
		@GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} go test ./...
		@echo " Tests OK."

$(EXEC): organize $(SOURCES)
		@echo "    Compilation des sources ${BUILD_TIME}"
		@GOPATH=$(PWD)/../.. GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o ${EXEC}-${VERSION} $(SOURCEDIR)/main.go
		@echo "    ${EXEC}-${VERSION} generated."

organize: audit
		@echo "    Go FMT"
		@$(foreach element,$(SOURCES),go fmt $(element);)

audit: deps
		@go tool vet -all .
		@echo "    Audit effectue"

deps: init
		@echo "    Download packages"
		@$(foreach element,$(PACKAGES),go get -d -v $(element);)

init: clean
		@echo "    Init of the project"
		@echo "    Version :: ${VERSION}"

execute: $(EXEC)
		./${EXEC}-${VERSION}

clean:
		@if [ -f "${EXEC}-${VERSION}" ] ; then rm ${EXEC}-${VERSION} ; fi
		@echo "    Nettoyage effectuee"


package-zip:  ${EXEC}
		@zip -r ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip ./${EXEC}-${VERSION}
		@echo "    Archive ${EXEC}-${GOOS}-${GOARCH}-${VERSION}.zip created"