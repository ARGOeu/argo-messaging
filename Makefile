PKGNAME=argo-messaging
SPECFILE=${PKGNAME}.spec
SHELL=bash
PKGVERSION = $(shell grep -s '^Version:' $(SPECFILE) | sed -e 's/Version: *//')
TMPDIR := $(shell mktemp -d /tmp/${PKGNAME}.XXXXXXXXXX)
GOPATH := $(shell mktemp -d /tmp/go.XXXXXXXXXX)
APPDIR := ${CURDIR}
GOFILES_NOVENDOR = $(shell go list ./... | grep -v '/vendor/' | sed -e 's/_\/usr\/src\/myapp/./g')

sources:
	mkdir -p ${TMPDIR}/${PKGNAME}-${PKGVERSION}/src/github.com/ARGOeu/argo-messaging
	cp -rp . ${TMPDIR}/${PKGNAME}-${PKGVERSION}/src/github.com/ARGOeu/argo-messaging
	cd ${TMPDIR} && tar czf ${PKGNAME}-${PKGVERSION}.tar.gz ${PKGNAME}-${PKGVERSION}
	mv ${TMPDIR}/${PKGNAME}-${PKGVERSION}.tar.gz .
	if [[ ${TMPDIR} == /tmp* ]]; then rm -rf ${TMPDIR} ;fi

go-build-linux-static:
	mkdir -p ${GOPATH}/src/github.com/ARGOeu/argo-messaging
	cp -R . ${GOPATH}/src/github.com/ARGOeu/argo-messaging
	cd ${GOPATH}/src/github.com/ARGOeu/argo-messaging && \
    export CGO_CFLAGS"=-O2 -fstack-protector --param=ssp-buffer-size=4 -D_FORTIFY_SOURCE=2"
	GOOS=linux go build -buildmode=pie -ldflags "-s -w -linkmode=external -extldflags '-z relro -z now'" -a -installsuffix cgo -o ${APPDIR}/argo-messaging-linux-static . &&\
	chown ${hostUID} ${APPDIR}/argo-messaging-linux-static

go-test:
	mkdir -p ${GOPATH}/src/github.com/ARGOeu/argo-messaging
	cp -R . ${GOPATH}/src/github.com/ARGOeu/argo-messaging
	cd ${GOPATH}/src/github.com/ARGOeu/argo-messaging && \
	go get github.com/axw/gocov/... && \
	go get github.com/AlekSi/gocov-xml && \
	${GOPATH}/bin/gocov test ${GOFILES_NOVENDOR} | ${GOPATH}/bin/gocov-xml > ${APPDIR}/coverage.xml &&\
	chown ${hostUID} ${APPDIR}/coverage.xml
