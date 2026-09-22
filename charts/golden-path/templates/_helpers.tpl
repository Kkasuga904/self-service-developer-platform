{{- define "golden-path.name" -}}
{{- .Values.metadata.name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "golden-path.labels" -}}
app.kubernetes.io/name: {{ include "golden-path.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
platform.example.io/owner: {{ .Values.spec.owner }}
platform.example.io/service: {{ .Values.metadata.name }}
platform.example.io/environment: {{ .Values.spec.environment }}
platform.example.io/contract-version: v1alpha1
{{- end }}

{{- define "golden-path.selectorLabels" -}}
app.kubernetes.io/name: {{ include "golden-path.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "golden-path.resources" -}}
{{- if eq .Values.spec.resources.size "small" }}
requests:
  cpu: 100m
  memory: 128Mi
limits:
  cpu: 500m
  memory: 256Mi
{{- else if eq .Values.spec.resources.size "medium" }}
requests:
  cpu: 250m
  memory: 256Mi
limits:
  cpu: "1"
  memory: 512Mi
{{- else if eq .Values.spec.resources.size "large" }}
requests:
  cpu: 500m
  memory: 512Mi
limits:
  cpu: "2"
  memory: 1Gi
{{- end }}
{{- end }}
