#!/bin/bash

FPM="${FPM:-fpm}"

die() {
    if [ $1 -eq 0 ]; then
        rm -rf "${TMPDIR}"
    else
        echo "Temporary data stored at '${TMPDIR}'" >&2
    fi
    echo "$2" >&2
    exit $1
}

GIT_VERSION="$(git describe --always --tags)" && {
    set -f; IFS='-' ; set -- ${GIT_VERSION}
    VERSION=$1; [ -z "$3" ] && RELEASE=$2 || RELEASE=$2.$3
    set +f; unset IFS

    [ "$RELEASE" == "" -a "$VERSION" != "" ] && RELEASE=0 

    if echo $VERSION | egrep '^v[0-9]+\.[0-9]+(\.[0-9]+)?$' >/dev/null; then
      VERSION=${VERSION:1:${#VERSION}}
      printf "'%s' '%s'\n" "$VERSION" "$RELEASE"
    fi
} || {
    exit 1
}

TMPDIR=$(mktemp -d)

DISTRO=$(lsb_release -i -s)
DRELEASE=$(lsb_release -r -s)

make || die 1 "Can't build package"
make DESTDIR="${TMPDIR}" install || die 1 "Can't install package"

# Determine if we are building for Ubuntu <15.04 and need to provide upstart script
is_upstart=0
if [[ "${DISTRO}" == "Ubuntu" ]]; then
	egrep -v -q '^(8|1[01234])\.' <<< ${DRELEASE}
	is_upstart=$?
fi

if [[ ${is_upstart} -eq 0 ]]; then
       mkdir -p "${TMPDIR}"/etc/systemd/system/
       mkdir -p "${TMPDIR}"/etc/default/
       cp ./contrib/carbonapi/deb/carbonapi.service "${TMPDIR}"/etc/systemd/system/ || die 1 "Copy error"
       cp ./contrib/carbonapi/common/carbonapi.env "${TMPDIR}"/etc/default/carbonapi || die 1 "Copy error"
else
       mkdir -p "${TMPDIR}"/etc/init/
       cp ./contrib/carbonapi/deb/carbonapi.conf "${TMPDIR}"/etc/init/ || die 1 "Copy error"
fi

mkdir -p "${TMPDIR}"/var/log/carbonapi/
mkdir -p "${TMPDIR}"/etc/logrotate.d/
cp ./contrib/carbonapi/deb/carbonapi.logrotate "${TMPDIR}"/etc/logrotate.d/carbonapi || die 1 "Copy error"

${FPM} -s dir -t deb -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p carbonapi_VERSION-ITERATION.ARCH.deb \
    -d "libcairo2 > 1.11" \
    --no-deb-systemd-restart-after-upgrade \
    --after-install contrib/carbonapi/fpm/systemd-reload.sh \
    --description "carbonapi: replacement graphite API server" \
    --license BSD-2 \
    --url "https://github.com/go-graphite/carbonapi" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
