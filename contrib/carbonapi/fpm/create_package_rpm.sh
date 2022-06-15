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

make || die 1 "Can't build package"
make DESTDIR="${TMPDIR}" install || die 1 "Can't install package"
mkdir -p "${TMPDIR}"/etc/systemd/system/
mkdir -p "${TMPDIR}"/etc/sysconfig/
cp ./contrib/carbonapi/rhel/carbonapi.service "${TMPDIR}"/etc/systemd/system/ || die 1 "Copy error"
cp ./contrib/carbonapi/common/carbonapi.env "${TMPDIR}"/etc/sysconfig/carbonapi || die 1 "Copy error"

${FPM} -s dir -t rpm -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p carbonapi-VERSION-ITERATION.ARCH.rpm \
    -d "cairo" \
    --after-install contrib/carbonapi/fpm/systemd-reload.sh \
    --description "carbonapi: replacement graphite API server" \
    --license BSD-2 \
    --url "https://github.com/grafana/carbonapi" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
