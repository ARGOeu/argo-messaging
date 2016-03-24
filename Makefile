PKGNAME=argo-messaging
SPECFILE=${PKGNAME}.spec
SHELL=bash
PKGVERSION = $(shell grep -s '^Version:' $(SPECFILE) | sed -e 's/Version: *//')
TMPDIR := $(shell mktemp -d /tmp/${PKGNAME}.XXXXXXXXXX)

sources:
	mkdir -p ${TMPDIR}/${PKGNAME}-${PKGVERSION}/src/github.com/ARGOeu/argo-messaging
	cp -rp . ${TMPDIR}/${PKGNAME}-${PKGVERSION}/src/github.com/ARGOeu/argo-messaging
	cd ${TMPDIR} && tar czf ${PKGNAME}-${PKGVERSION}.tar.gz ${PKGNAME}-${PKGVERSION}
	mv ${TMPDIR}/${PKGNAME}-${PKGVERSION}.tar.gz .
	if [[ ${TMPDIR} == /tmp* ]]; then rm -rf ${TMPDIR} ;fi
