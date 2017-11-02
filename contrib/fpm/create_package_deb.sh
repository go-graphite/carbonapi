#!/bin/bash -x
die() {
    if [[ $1 -eq 0 ]]; then
        rm -rf "${TMPDIR}"
    else
        echo "Temporary data stored at '${TMPDIR}'"
    fi
    echo "$2"
    exit $1
}

pwd
git fetch --tags
VERSION=$(git describe --abbrev=6 --always --tags)
TMPDIR=$(mktemp -d)

DISTRO=$(lsb_release -i -s)
RELEASE=$(lsb_release -r -s)

echo "version: ${VERSION}"

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
       cp ./contrib/deb/carbonapi.service "${TMPDIR}"/etc/systemd/system/
       cp ./contrib/common/carbonapi.env "${TMPDIR}"/etc/default/carbonapi
else
       mkdir -p "${TMPDIR}"/etc/init/
       cp ./contrib/deb/carbonapi.conf "${TMPDIR}"/etc/init/
fi

fpm -s dir -t deb -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    -p carbonapi_VERSION_ARCH.deb \
    -d "libcairo2 > 1.11" \
    --no-deb-systemd-restart-after-upgrade \
    --after-install contrib/fpm/systemd-reload.sh \
    --description "carbonapi: replacement graphite API server" \
    --license BSD-2 \
    --url "https://github.com/go-graphite/" \
    "${@}" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
