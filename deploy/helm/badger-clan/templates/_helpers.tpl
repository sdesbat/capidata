{{- define "badger-clan.name" -}}{{ .Chart.Name }}{{- end -}}
{{- define "badger-clan.fullname" -}}{{ include "badger-clan.name" . }}{{- end -}}
{{- define "badger-clan.serviceAccountName" -}}{{ if .Values.serviceAccount.name }}{{ .Values.serviceAccount.name }}{{ else }}{{ include "badger-clan.fullname" . }}{{ end }}{{- end -}}
