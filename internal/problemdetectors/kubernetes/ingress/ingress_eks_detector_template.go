package ingress

const (
	InvalidAnnotationSolution = `
1. Please update below invalid annotation with ingress {{ .ingressName }}.

{{ .annotations }}

2. Use below commands to check ingress annotations.
   kubectl get ingress {{ .ingressName }} -o yaml -n {{ .namespace }}
`
	SSLAnnotationSolution = `
Please check if SSL related annotation {{ .sslAnnotation }} is added in ingress {{ .ingressName }}.
`
	InvalidServiceSolution = `
1. Check below rules with ingress {{ .ingressName }}.

{{ .rules }}
`
	InvalidServiceStep2Solution = `
2. Use below commands to check ingress rules and compare with service.
   kubectl get ingress {{ .ingressName }} -o yaml -n {{ .namespace }}
   kubectl get service -n {{ .namespace }}
`
	ALBNotCreatedSolution = `
Please check if application load balancer {{ .alb }} is created in AWS.
`
	UnhealthTargetGroupSolution = `
Please check below instance health under EC2 -> Load Balancing -> Target Groups in AWS.

{{ .targets }}
`
	InvalidCertificateSolution = `
Please check if below certificate exist and valid in AWS.

{{ .certificates }}
`
	InvalidCertificateHostSolution = `
Please check if below domain name in certificate is valid in AWS.

Ingress host name: {{ .hostName}}
Host name(s) in certificate:
{{ .certificates }}
`
)
