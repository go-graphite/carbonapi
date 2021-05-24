#!/usr/bin/env bash

FPM="${FPM:-fpm}"

VERSION=$(git describe --abbrev=4 --always --tags)
TMPDIR=$(mktemp -d)

DISTRO=$(lsb_release -i -s)
RELEASE=$(lsb_release -r -s)
NAME="carbonzipper"

die() {
    if [ $1 -eq 0 ]; then
        rm -rf "${TMPDIR}"
    else
        echo "Temporary data stored at '${TMPDIR}'"  >&2
    fi
    echo "$2" >&2
    exit $1
}

make || die 1 "Can't build package"
make DESTDIR="${TMPDIR}" install || die 1 "Can't install package"

# Determine if we are building for Ubuntu <15.04 and need to provide upstart script
is_upstart=0
if [[ "${DISTRO}" == "Ubuntu" ]]; then
	egrep -v -q '^(8|1[01234])\.' <<< ${RELEASE}
	is_upstart=$?
fi

if [[ ${is_upstart} -eq 0 ]]; then
       mkdir -p "${TMPDIR}"/etc/systemd/system/
       mkdir -p "${TMPDIR}"/etc/default/
       cp ./contrib/carbonzipper/deb/${NAME}.service "${TMPDIR}"/etc/systemd/system/ || die 1 "Copy error"
       cp ./contrib/carbonzipper/common/${NAME}.env "${TMPDIR}"/etc/default/${NAME} || die 1 "Copy error"
else
       mkdir -p "${TMPDIR}"/etc/init/
       cp ./contrib/carbonzipper/deb/${NAME}.conf "${TMPDIR}"/etc/init/ || die 1 "Copy error"
fi

${FPM} -s dir -t deb -n ${NAME} -v ${VERSION} -C ${TMPDIR} \
    -p ${NAME}_VERSION_ARCH.deb \
    --no-deb-systemd-restart-after-upgrade \
    --after-install contrib/carbonzipper/fpm/systemd-reload.sh \
    --description "carbonzipper proxy for graphite-web and carbonapi" \
    --license MIT \
    --url "https://github.com/go-graphite/carbonapi" \
    "${@}" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
