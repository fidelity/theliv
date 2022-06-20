# Default Prometheus Alerts
Below are the default Prometheus alerts provided by Theliv.   

### Pod Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| PodNotRunning | `kube_pod_status_phase{job='kube-state-metrics',phase=~'Failed|Pending|Unknown'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0 ` |
| PodNotReady | `kube_pod_status_phase{job='kube-state-metrics',phase='Running'} * on(uid) group_left(condition) kube_pod_status_ready{job='kube-state-metrics', condition='false'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |  

### Container Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| ContainerWaitingAsCrashLoopBackoff | `kube_pod_container_status_waiting_reason{job='kube-state-metrics', reason='CrashLoopBackOff'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerWaitingAsImagePullBackOff | `kube_pod_container_status_waiting_reason{job='kube-state-metrics', reason=~'ImagePullBackOff|ErrImagePull|InvalidImageName'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerWaitingAsCreateContainerError | `kube_pod_container_status_waiting_reason{job='kube-state-metrics', reason=~'CreateContainerConfigError|CreateContainerError'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerTerminatedAsOOMKilled | `kube_pod_container_status_terminated_reason{job='kube-state-metrics',reason='OOMKilled'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerTerminatedAsError | `kube_pod_container_status_terminated_reason{job='kube-state-metrics',reason='Error'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerTerminatedAsContainerCannotRun | `kube_pod_container_status_terminated_reason{job='kube-state-metrics',reason='ContainerCannotRun'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerTerminatedAsDeadlineExceeded | `kube_pod_container_status_terminated_reason{job='kube-state-metrics',reason='DeadlineExceeded'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| ContainerTerminatedAsEvicted | `kube_pod_container_status_terminated_reason{job='kube-state-metrics',reason='Evicted'} * on(uid, container) group_left(image) kube_pod_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |  

### InitContainer Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| InitContainerWaitingAsCrashLoopBackoff | `kube_pod_init_container_status_waiting_reason{job='kube-state-metrics', reason='CrashLoopBackOff'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerWaitingAsImagePullBackOff | `kube_pod_init_container_status_waiting_reason{job='kube-state-metrics', reason=~'ImagePullBackOff|ErrImagePull|InvalidImageName'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerTerminatedAsOOMKilled | `kube_pod_init_container_status_terminated_reason{job='kube-state-metrics',reason='OOMKilled'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerTerminatedAsError | `kube_pod_init_container_status_terminated_reason{job='kube-state-metrics',reason='Error'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerTerminatedAsContainerCannotRun | `kube_pod_init_container_status_terminated_reason{job='kube-state-metrics',reason='ContainerCannotRun'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerTerminatedAsDeadlineExceeded | `kube_pod_init_container_status_terminated_reason{job='kube-state-metrics',reason='DeadlineExceeded'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |
| InitContainerTerminatedAsEvicted | `kube_pod_init_container_status_terminated_reason{job='kube-state-metrics',reason='Evicted'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0` |  

### Deployment Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| DeploymentNotAvailable | `kube_deployment_status_condition{job='kube-state-metrics', condition='Available', status!='true'} >0` |
| DeploymentGenerationMismatch | `kube_deployment_status_observed_generation{job='kube-state-metrics'} - on (deployment, namespace) kube_deployment_metadata_generation{job='kube-state-metrics'} !=0` |
| DeploymentReplicasMismatch | `(kube_deployment_spec_replicas{job='kube-state-metrics'} - on(deployment, namespace) kube_deployment_status_replicas_available{job='kube-state-metrics'} !=0 ) and (kube_deployment_status_replicas_updated ==0)` |  

### Node Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| NodeNotReady | `kube_node_status_condition{job='kube-state-metrics', condition!='Ready', status=~'true|unknown'} >0` |
| NodeDiskPressure | `kube_node_status_condition{job='kube-state-metrics', condition='DiskPressure', status='true'} >0` |
| NodeMemoryPressure | `kube_node_status_condition{job='kube-state-metrics', condition='MemoryPressure', status='true'} >0` |
| NodePIDPressure | `kube_node_status_condition{job='kube-state-metrics', condition='PIDPressure', status='true'} >0` |
| NodeNetworkUnavailable | `kube_node_status_condition{job='kube-state-metrics', condition='NetworkUnavailable', status='true'} >0` |   

### Service Endpoints Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| EndpointAddressNotAvailable | `kube_endpoint_address_available{job='kube-state-metrics'} == 0` |   

### StatefulSet Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| StatefulsetGenerationMismatch | `kube_statefulset_status_observed_generation{job='kube-state-metrics'} - on (statefulset, namespace) kube_statefulset_metadata_generation{job='kube-state-metrics'} !=0` |
| StatefulsetReplicasMismatch | `(kube_statefulset_replicas{job='kube-state-metrics'} - on(deployment, namespace) kube_statefulset_status_replicas_ready{job='kube-state-metrics'} !=0 ) and (kube_statefulset_status_replicas_updated{job='kube-state-metrics'} ==0)` |
| StatefulsetUpdateNotRolledOut | `(max without (revision) (kube_statefulset_status_current_revision{job='kube-state-metrics'} unless kube_statefulset_status_update_revision{job='kube-state-metrics'}) *  (kube_statefulset_replicas{job='kube-state-metrics'} != kube_statefulset_status_replicas_updated{job='kube-state-metrics'})) and (changes(kube_statefulset_status_replicas_updated{job='kube-state-metrics'}[5m]) == 0)` |   

### DaemonSet Alerts
| Alert Name | Alert Expression (PromQL) |
| ----------- | ----------- |
| DaemonSetRolloutStuck | `((kube_daemonset_status_current_number_scheduled{job='kube-state-metrics'}!=kube_daemonset_status_desired_number_scheduled{job='kube-state-metrics'}) or (kube_daemonset_status_number_misscheduled{job='kube-state-metrics'}!=0) or (kube_daemonset_status_updated_number_scheduled{job='kube-state-metrics'}!=kube_daemonset_status_desired_number_scheduled{job='kube-state-metrics'}) or (kube_daemonset_status_number_available{job='kube-state-metrics'}!=kube_daemonset_status_desired_number_scheduled{job='kube-state-metrics'})) and (changes(kube_daemonset_status_updated_number_scheduled{job='kube-state-metrics'}[5m])==0)` |
| DaemonSetNotScheduled | `kube_daemonset_status_desired_number_scheduled{job='kube-state-metrics'} - kube_daemonset_status_current_number_scheduled{job='kube-state-metrics'} > 0` |
| DaemonSetMissScheduled | `kube_daemonset_status_number_misscheduled{job='kube-state-metrics'}>0` |
| DaemonSetUnavailable | `kube_daemonset_status_number_unavailable{job='kube-state-metrics'} >0` |   