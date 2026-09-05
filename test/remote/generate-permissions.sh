#!/bin/sh

BASEPATH=$(dirname "$0")
PERMFILE=${BASEPATH}/cd-remote-permissions.yaml
if ! test -f "${PERMFILE}"; then
	echo "ERROR: $PERMFILE does not exist." >&2
	exit 1
fi

NAMESPACE=${NAMESPACE:-cd-e2e}

if test "${CD_E2E_NAME_PREFIX}" != ""; then
	CRNAME="${CD_E2E_NAME_PREFIX}-hanzocd-application-controller"
	CRBNAME="${CD_E2E_NAME_PREFIX}-hanzocd-application-controller"
	CONTROLLERSANAME="${CD_E2E_NAME_PREFIX}-hanzocd-application-controller"
	SERVERSANAME="${CD_E2E_NAME_PREFIX}-hanzocd-server"
else
	CRNAME="hanzocd-application-controller"
	CRBNAME="hanzocd-application-controller"
	CONTROLLERSANAME="hanzocd-application-controller"
	SERVERSANAME="hanzocd-server"
fi

sed \
	-e "s/##CRNAME##/${CRNAME}/g" \
	-e "s/##CRBNAME##/${CRBNAME}/g" \
	-e "s/##CONTROLLERSANAME##/${CONTROLLERSANAME}/g" \
	-e "s/##SERVERSANAME##/${SERVERSANAME}/g" \
	-e "s/##NAMESPACE##/${NAMESPACE}/g" \
	"$PERMFILE"
