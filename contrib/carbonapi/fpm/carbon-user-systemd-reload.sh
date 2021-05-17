#!/bin/sh

# based on https://github.com/leoleovich/grafsy/blob/master/packaging/postinst

# Abort if any command returns an error value
set -e

USER=carbon
GROUP=carbon
CONF=/etc/carbonapi/carbonapi.yaml
PIDFILE=/var/run/carbonapi/carbonapi.pid
LOGFILE=/var/log/carbonapi/carbonapi.log

# Following user part should be tested on both RPM and DEB systems
if ! getent group "${GROUP}" > /dev/null 2>&1 ; then
  groupadd --system "${GROUP}"
fi
GID=$(getent group "${GROUP}" | cut -d: -f 3)
if ! id "${USER}" > /dev/null 2>&1 ; then
  adduser --system --home /dev/null --no-create-home \
    --gid "${GID}" --shell /bin/false \
    "${USER}"
fi

# fix PID permissions
if [ -f "${PIDFILE}" ]; then
    chown "${USER}":"${GROUP}" "${PIDFILE}"
fi

# fix log permissions
if [ -f "${LOGFILE}" ]; then
    chown "${USER}":"${GROUP}" "${LOGFILE}"
fi

if [ ! -e "${CONF}" ]; then
  echo "For use this software you have to create ${CONF} file. You could use /usr/share/carbonapi/carbonapi.example.yaml as default"
else
  # On debian jessie (systemd 215) it fails if symlink already exists
  systemctl is-enabled carbonapi || systemctl enable carbonapi
  # Check if systemd is up and running, e.g. not in chroot
  if systemctl 1>/dev/null 2>&1; then
    systemctl daemon-reload
    systemctl restart carbonapi.service
  fi
fi