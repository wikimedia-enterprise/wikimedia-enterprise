#!/bin/sh

CUSTOM_ROOT_CA_TMP_FILE="/tmp/tmp-root-ca.crt"

if [ -z "${INTERNAL_ROOT_CA_PEM}" ]; then
    echo "INTERNAL_ROOT_CA_PEM is not set"
else
    echo -e "${INTERNAL_ROOT_CA_PEM}" > ${CUSTOM_ROOT_CA_TMP_FILE}
    sudo cp ${CUSTOM_ROOT_CA_TMP_FILE} /usr/local/share/ca-certificates/wme-internal-root-ca.crt
    sudo update-ca-certificates
    if [ -f ${CUSTOM_ROOT_CA_TMP_FILE} ]; then rm -f ${CUSTOM_ROOT_CA_TMP_FILE}; fi
fi

./main
