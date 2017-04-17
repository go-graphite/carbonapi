#!/usr/bin/env bash
VERSION=$(git describe --abbrev=4 --dirty --always --tags)
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

fpm -s dir -t deb -n carbonapi -v ${VERSION} -C ${TMPDIR} \
  -p carbonapi_VERSION_ARCH.deb \
  -d "libcairo2 > 1.11" \
  usr/bin usr/share || die "Can't create package!"

die 0 "Success"
