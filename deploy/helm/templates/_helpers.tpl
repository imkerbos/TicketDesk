{{/*
通用标签
*/}}
{{- define "ticketdesk.labels" -}}
app.kubernetes.io/name: ticketdesk
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
MySQL 连接地址
*/}}
{{- define "ticketdesk.mysql.host" -}}
{{- if .Values.mysql.enabled -}}
{{ .Release.Name }}-mysql
{{- else -}}
{{ .Values.mysql.external.host }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.mysql.port" -}}
{{- if .Values.mysql.enabled -}}
{{ .Values.mysql.port }}
{{- else -}}
{{ .Values.mysql.external.port }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.mysql.user" -}}
{{- if .Values.mysql.enabled -}}
{{ .Values.mysql.user }}
{{- else -}}
{{ .Values.mysql.external.user }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.mysql.password" -}}
{{- if .Values.mysql.enabled -}}
{{ .Values.mysql.password }}
{{- else -}}
{{ .Values.mysql.external.password }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.mysql.database" -}}
{{- if .Values.mysql.enabled -}}
{{ .Values.mysql.database }}
{{- else -}}
{{ .Values.mysql.external.database }}
{{- end -}}
{{- end }}

{{/*
Redis 连接地址
*/}}
{{- define "ticketdesk.redis.host" -}}
{{- if .Values.redis.enabled -}}
{{ .Release.Name }}-redis
{{- else -}}
{{ .Values.redis.external.host }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.redis.port" -}}
{{- if .Values.redis.enabled -}}
{{ .Values.redis.port }}
{{- else -}}
{{ .Values.redis.external.port }}
{{- end -}}
{{- end }}

{{- define "ticketdesk.redis.password" -}}
{{- if .Values.redis.enabled -}}
{{ .Values.redis.password }}
{{- else -}}
{{ .Values.redis.external.password }}
{{- end -}}
{{- end }}
