#!/usr/bin/env bash
VERSION_GIT=$(git describe --abbrev=4 --always --tags | rev | sed 's/-/./' | rev) 
VERSION=$(echo ${VERSION_GIT} | cut -d'-' -f 1)
RELEASE=$(echo ${VERSION_GIT} | cut -d'-' -f 2)
TMPDIR=$(mktemp -d)
NAME="carbonzipper"

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
cp ./contrib/rhel/${NAME}.service "${TMPDIR}"/etc/systemd/system/
cp ./contrib/common/${NAME}.env "${TMPDIR}"/etc/sysconfig/${NAME}

fpm -s dir -t rpm -n ${NAME} -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p ${NAME}_VERSION-ITERATION_ARCH.rpm \
    -d "cairo > 1.11" \
    --after-install contrib/fpm/systemd-reload.sh \
    --description "carbonserver proxy for graphite-web and carbonapi" \
    --license MIT \
    --url "https://github.com/go-graphite/" \
    "${@}" \
    etc usr/bin usr/share || die "Can't create package!"

die 0 "Success"
