#!/bin/sh

BASEPATH=$(dirname "$0")
PERMFILE=${BASEPATH}/argocd-remote-permissions.yaml
if ! test -f "${PERMFILE}"; then
	echo "ERROR: $PERMFILE does not exist." >&2
	exit 1
fi

NAMESPACE=${NAMESPACE:-argocd-e2e}

if test "${CD_E2E_NAME_PREFIX}" != ""; then
	CRNAME="${CD_E2E_NAME_PREFIX}-argocd-application-controller"
	CRBNAME="${CD_E2E_NAME_PREFIX}-argocd-application-controller"
	CONTROLLERSANAME="${CD_E2E_NAME_PREFIX}-argocd-application-controller"
	SERVERSANAME="${CD_E2E_NAME_PREFIX}-argocd-server"
else
	CRNAME="argocd-application-controller"
	CRBNAME="argocd-application-controller"
	CONTROLLERSANAME="argocd-application-controller"
	SERVERSANAME="argocd-server"
fi

sed \
	-e "s/##CRNAME##/${CRNAME}/g" \
	-e "s/##CRBNAME##/${CRBNAME}/g" \
	-e "s/##CONTROLLERSANAME##/${CONTROLLERSANAME}/g" \
	-e "s/##SERVERSANAME##/${SERVERSANAME}/g" \
	-e "s/##NAMESPACE##/${NAMESPACE}/g" \
	"$PERMFILE"
