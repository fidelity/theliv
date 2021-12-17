package kubernetes

const (
	UnSchedulableSolution = `
1. Node {{ .ObjectMeta.Name }} is unschedulable, due to your node is cordoned.
2. Cmd to describe Node: kubectl describe node {{ .ObjectMeta.Name }}.
3. Cmd to uncordon the Node: kubectl uncordon {{ .ObjectMeta.Name }}.
4. If keeping Node cordoned is necessary, please ensure schedulable Nodes exist in Cluster.
5. Cmd to check Node Taints: kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints
`

	NotReadySolution = `
1. Node {{ .ObjectMeta.Name }} is not Ready.
2. NotReady message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "Ready" }}{{$value.Message}}{{- end}}{{end}}.
3. Please wait for sometime, if issue persists, you may need to ssh to the Node and check.
4. Cmd to check Kubelet log: journalctl -u kubelet.
5. Cmd to check Kubelet service status: systemctl status kubelet.
6. Cmd to restart Kubelet service: systemctl restart kubelet.
`

	UnReachableSolution = `
1. Node {{ .ObjectMeta.Name }} is UnReachable.
2. Please check if Kubelet service in node is started, and there's no network connection issue.
3. You may need to ssh to the Node and check the Kubelet service.
4. Cmd to check Kubelet log: journalctl -u kubelet.
5. Cmd to check Kubelet service status: systemctl status kubelet.
6. Cmd to restart Kubelet service: systemctl restart kubelet.
`

	MemPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting memory pressure issue.
2. MemoryPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "MemoryPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, kill unnecessary processes.
4. You can find Pods running on this Node, re-allocate some to other Nodes.
5. Cmd to find Pods on this Node: kubectl get pods -o wide -A | grep {{ .ObjectMeta.Name }}
`

	DiskPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting disk pressure issue.
2. DiskPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "DiskPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, delete unnecessary files, or add more storage resources.
`

	PidPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting Pid pressure issue.
2. PIDPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "PIDPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, kill unnecessary processes.
4. You can find Pods running on this Node, re-allocate some to other Nodes.
5. Cmd to find Pods on this Node: kubectl get pods -o wide -A | grep {{ .ObjectMeta.Name }}
`

	NetUnAvailableSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting network unavailable issue.
2. NetworkUnavailable message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "NetworkUnavailable" }}{{$value.Message}}{{- end}}{{end}}.
3. Please check if Kubelet service in node is started, and there's no network connection issue.
4. You may need to ssh to the Node and check the Kubelet service.
5. Cmd to check Kubelet log: journalctl -u kubelet.
6. Cmd to check Kubelet service status: systemctl status kubelet.
7. Cmd to restart Kubelet service: systemctl restart kubelet.
`

	UnInitializedSolution = `
1. Node {{ .ObjectMeta.Name }} is uninitialized yet by the external cloud provider.
2. Please wait for sometime and check the Node status again.
3. If issue persists, please check if the cloud-controller-manager correctly initialized this Node.
4. Cmd to describe Node: kubectl describe node {{ .ObjectMeta.Name }}.
`

	GeneralErrSolution = `
1. Node {{ .ObjectMeta.Name }} is not Ready or is unschedulable.
2. Please check if your node is started, not cordoned, and the Kubelet running well, no network connection issues.
3. Please ensure schedulable Nodes exist in Cluster.
4. Cmd to describe Node: kubectl describe node {{ .ObjectMeta.Name }}.
`
)
