#!/usr/bin/env bash
VERSION_GIT=$(git describe --abbrev=4 --always --tags | rev | sed 's/-/./' | rev) 
VERSION=$(cut -d'-' -f 1 <<< ${VERSION_GIT})
RELEASE=$(cut -d'-' -f 2 <<< ${VERSION_GIT})
if [[ "${VERSION}" == "${RELEASE}" ]]; then
       RELEASE="1"
else
       REL_VERSION=$(cut -d'.' -f 1 <<< ${RELEASE})
       REL_COMMIT=$(cut -d'.' -f 2 <<< ${RELEASE})
       RELEASE="$((REL_VERSION+1)).${REL_COMMIT}"
fi
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
cp ./contrib/rhel/carbonapi.service "${TMPDIR}"/etc/systemd/system/
cp ./contrib/common/carbonapi.env "${TMPDIR}"/etc/sysconfig/carbonapi

fpm -s dir -t rpm -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p carbonapi_VERSION-ITERATION_ARCH.rpm \
    -d "cairo" \
    --after-install contrib/fpm/systemd-reload.sh \
    --description "carbonapi: replacement graphite API server" \
    --license MIT \
    --url "https://github.com/go-graphite/" \
    "${@}" \
    etc usr/bin usr/share || die "Can't create package!"

die 0 "Success"
