#!/usr/bin/env bash
VERSION=$(git describe --abbrev=4 --always --tags)
TMPDIR=$(mktemp -d)

die() {
    if [[ $1 -eq 0 ]]; then
        rm -rf "${TMPDIR}"
    else
        echo "Temporary data stored at '${TMPDIR}'"
    fi
    echo "$2"
    exit $1
}

make || die 1 "Can't build package"
make DESTDIR="${TMPDIR}" install || die 1 "Can't install package"
mkdir -p "${TMPDIR}"/etc/systemd/system/
mkdir -p "${TMPDIR}"/etc/default/
cp contrib/deb/carbonapi.service "${TMPDIR}"/etc/systemd/system/
cp contrib/deb/carbonapi "${TMPDIR}"/etc/default/

fpm -s dir -t deb -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    -p carbonapi_VERSION_ARCH.deb \
    -d "libcairo2 > 1.11" \
    --no-deb-systemd-restart-after-upgrade \
    --after-install contrib/fpm/systemd-reload.sh \
    usr/bin usr/share || die "Can't create package!"

die 0 "Success"
