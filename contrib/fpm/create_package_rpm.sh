#!/usr/bin/env bash
VERSION_GIT=$(git describe --abbrev=4 --always --tags | rev | sed 's/-/./' | rev) 
VERSION=$(echo ${VERSION_GIT} | cut -d'-' -f 1)
RELEASE=$(echo ${VERSION_GIT} | cut -d'-' -f 2)
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
mkdir -p "${TMPDIR}"/etc/sysconfig/
cp contrib/rhel/carbonapi.service "${TMPDIR}"/etc/systemd/system/
cp contrib/rhel/carbonapi "${TMPDIR}"/etc/sysconfig/

fpm -s dir -t rpm -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p carbonapi_VERSION-ITERATION_ARCH.deb \
    -d "cairo > 1.11" \
    --after-install contrib/fpm/systemd-reload.sh \
    usr/bin usr/share || die "Can't create package!"

die 0 "Success"
