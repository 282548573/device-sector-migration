INSTALLDIR:="device-sector-migration"

all: build pack
.PHONY: build pack

build:
	rm -rf build/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sector-migration-tool main.go

pack:
	rm -rf ${INSTALLDIR}
	mkdir -p ${INSTALLDIR}
	mkdir -p ${INSTALLDIR}/configs
	cp -rf ./config.yaml ./sectors.json ${INSTALLDIR}/configs
	cp ./sector-migration-tool ${INSTALLDIR}
	tar -czf ./${INSTALLDIR}.tar.gz ${INSTALLDIR}

