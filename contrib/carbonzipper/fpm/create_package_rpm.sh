#!/usr/bin/env bash

FPM="${FPM:-fpm}"

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

NAME="carbonzipper"

die() {
    if [ $1 -eq 0 ]; then
        rm -rf "${TMPDIR}"
    else
        echo "Temporary data stored at '${TMPDIR}'" >&2
    fi
    echo "$2" >&2
    exit $1
}

MAJOR_DISTRO_VERSION=$(lsb_release -s -r | sed -e 's/\..*//')
[ "${MAJOR_DISTRO_VERSION}" == "" ] && MAJOR_DISTRO_VERSION="7"

TMPDIR=$(mktemp -d)

make || die 1 "Can't build package"
make DESTDIR="${TMPDIR}" install || die 1 "Can't install package"
mkdir -p "${TMPDIR}"/etc/sysconfig/
cp ./contrib/carbonzipper/common/${NAME}.env "${TMPDIR}"/etc/sysconfig/${NAME} || dir 1 "Copy error"
if [[ "${MAJOR_DISTRO_VERSION}" -le 6 ]]; then
	mkdir -p "${TMPDIR}"/init.d
	cp ./contrib/carbonzipper/rhel/${NAME}.init "${TMPDIR}"/etc/init.d/${NAME} || dir 1 "Copy error"
else
	mkdir -p "${TMPDIR}"/etc/systemd/system/
	cp ./contrib/carbonzipper/rhel/${NAME}.service "${TMPDIR}"/etc/systemd/system/ || dir 1 "Copy error"
fi


${FPM} -s dir -t rpm -n ${NAME} -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p ${NAME}-VERSION-ITERATION.ARCH.rpm \
    --after-install contrib/carbonzipper/fpm/systemd-reload.sh \
    --description "carbonzipper proxy for graphite-web and carbonapi" \
    --license MIT \
    --url "https://github.com/go-graphite/carbonapi" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
