package kubernetes

const (
	PendingNodeSelectorSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) didn't match node selector.
2. Please check the Pod node-selector, affinity.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
4. Cmd to check Node Labels: kubectl get node --show-labels
5. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
`
	PendingNodeTaintSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) had taints, that the pod didn't tolerate.
2. Please check the each Node's taints, and modify the pod to tolerate 1 of them.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
4. Cmd to check Node Taints: kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers
5. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
`

	PendingNodeUnschedulableSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) were unschedulable.
2. Please check the Nodes STATUS, if all nodes Not Ready or SchedulingDisabled.
3. To fix this, you may need to fix the issue in NotReady Node, or uncordon the Node, or add 1 new Node.
4. Reference link: https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/
5. Cmd to check Node: kubectl get nodes
6. Cmd to uncordon a Node: kubectl uncordon <node name>
`

	PVCNotFoundSolution = "1. Pod {{ .ObjectMeta.Name }} is pending, used PVC not found." + KubectlPodAndPVC

	PVCUnboundSolution = "1. Pod {{ .ObjectMeta.Name }} is pending, due to use an unbound PVC." + KubectlPodAndPVC

	KubectlPodAndPVC = `
2. Please check PVC used by the pod.
3. Reference link: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
4. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
5. Cmd to check Pvc: kubectl get pvc -n {{ .ObjectMeta.Namespace }}
`

	PendingInsufficientSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) has insufficient resources.
2. Please check the resources that the pod requests or limits, try to modify them to apllicable quota.
3. Reference link: https://kubernetes.io/docs/concepts/architecture/nodes/
4. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
5. Cmd to check Node allocatable: kubectl get nodes -o custom-columns=NAME:.metadata.name,ALLOCATABLE:.status.allocatable --no-headers
`

	PendingNodeAffinitySolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) didn't match pod affinity/anti-affinity rules. Also some Nodes may include taints that the Pod didn't tolerate.
2. Please check the Pod affinity/anti-affinity. Or check the Nodes' taints, and make the Pod tolerate 1 of them.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity
4. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
5. Cmd to check Node Labels: kubectl get nodes --show-labels
6. Cmd to check Node Taints: kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers
`

	PendingNoHostPortSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) didn't have free ports for the requested pod ports.
2. Please check the HostPort used in the Pod, change/remove it is suggested.
3. If the HostPort is necessary, use NodeSelector to assign Pod to the Node which the specified Port is available.
4. Cmd to check Pod: kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} -o yaml
5. Cmd to check Node Labels: kubectl get nodes --show-labels
6. Cmd to check Node Taints: kubectl get nodes -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers
`
)
