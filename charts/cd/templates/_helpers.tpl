{{/*
Image reference shared by every hanzocd component (dex and redis-ha pin
their own upstream images and do not use this helper).
*/}}
{{- define "cd.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) -}}
{{- end -}}

{{/*
Bookkeeping labels added on top of each resource's own
app.kubernetes.io/{name,part-of,component} labels. Never used in a
selector: those must stay byte-identical to the name every component
already resolves at runtime.
*/}}
{{- define "cd.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Effective server replica count: explicit value wins, otherwise 2 under
ha.enabled and 1 otherwise.
*/}}
{{- define "cd.server.replicaCount" -}}
{{- .Values.server.replicaCount | default (ternary 2 1 .Values.ha.enabled) -}}
{{- end -}}

{{/*
Effective repo-server replica count: same rule as the server.
*/}}
{{- define "cd.repoServer.replicaCount" -}}
{{- .Values.repoServer.replicaCount | default (ternary 2 1 .Values.ha.enabled) -}}
{{- end -}}
