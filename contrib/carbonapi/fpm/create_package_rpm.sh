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

pwd
VERSION_GIT=$(git describe --abbrev=6 --always --tags | rev | sed 's/-/./' | rev)
VERSION=$(cut -d'-' -f 1 <<< ${VERSION_GIT})
RELEASE=$(cut -d'-' -f 2 <<< ${VERSION_GIT})
COMMIT=$(cut -d'-' -f 3 <<< ${VERSION_GIT})

REL_VERSION=""
REL_COMMIT=""
if [[ "${VERSION}" == "${RELEASE}" ]]; then
       RELEASE="1"
else
       REL_VERSION=$(cut -d'.' -f 1 <<< ${RELEASE})
       REL_COMMIT=$(cut -d'.' -f 2 <<< ${RELEASE})
       if [[ ! -z "${COMMIT}" ]]; then
           REL_COMMIT=${COMMIT}
       fi
       case "${REL_VERSION}" in
           "beta")
               RELEASE="0.2.${RELEASE/./}"
               if [[ ! -z "${REL_COMMIT}" ]]; then
                   RELEASE="${RELEASE}.${REL_COMMIT}"
               fi
               ;;
           "rc")
               RELEASE="0.3.${RELEASE/./}"
               if [[ ! -z "${REL_COMMIT}" ]]; then
                   RELEASE="${RELEASE}.${REL_COMMIT}"
               fi
               ;;
           *)
               RELEASE="1.0.post$((REL_COMMIT+1))"
               ;;
       esac
fi
grep '^[0-9]\+\.[0-9]\+\.' <<< ${VERSION} || {
	echo "Revision: $(git rev-parse HEAD)";
	echo "Version: $(git describe --abbrev=6 --always --tags)";
	echo "Known tags: $(git tag)";
	echo;
	echo;
	die 1 "Can't get latest version from git";
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
    --url "https://github.com/go-graphite/carbonapi" \
    "${@}" \
    etc usr/bin usr/share || die 1 "Can't create package!"

die 0 "Success"
