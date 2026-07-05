{{/*
Base name for all resources rendered by this chart.
*/}}
{{- define "calendar-chart.fullname" -}}
{{- .Release.Name -}}
{{- end -}}

{{/*
Common labels shared by every resource.
*/}}
{{- define "calendar-chart.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
