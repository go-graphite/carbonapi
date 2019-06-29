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
VERSION_GIT=$(git describe --abbrev=6 --always --tags | rev | sed 's/-/./' | rev)
VERSION=$(cut -d'-' -f 1 <<< ${VERSION_GIT})
RELEASE=$(cut -d'-' -f 2 <<< ${VERSION_GIT})
if [[ "${VERSION}" == "${RELEASE}" ]]; then
       RELEASE="1"
else
       REL_VERSION=$(cut -d'.' -f 1 <<< ${RELEASE})
       REL_COMMIT=$(cut -d'.' -f 2 <<< ${RELEASE})
       RELEASE="$((REL_VERSION+1)).${REL_COMMIT}"
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
cp ./contrib/carbonapi/rhel/carbonapi.service "${TMPDIR}"/etc/systemd/system/
cp ./contrib/carbonapi/common/carbonapi.env "${TMPDIR}"/etc/sysconfig/carbonapi

fpm -s dir -t rpm -n carbonapi -v ${VERSION} -C ${TMPDIR} \
    --iteration ${RELEASE} \
    -p carbonapi-VERSION-ITERATION.ARCH.rpm \
    -d "cairo" \
    --after-install contrib/carbonapi/fpm/systemd-reload.sh \
    --description "carbonapi: replacement graphite API server" \
    --license BSD-2 \
    --url "https://github.com/go-graphite/carbonapi" \
    "${@}" \
    etc usr/bin usr/share || die "Can't create package!"

die 0 "Success"
