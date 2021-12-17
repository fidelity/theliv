# Autodetected kubernetes issues

## This document contains the list of kubernetes related issues automatically detected by Theliv

## 1.ImagePullBackOff

| error_code    | status  | tags                     |
| ------------- | ------- | ------------------------ |
| IMAGEPULL_ERR | pending | imagepullbackoff,kubelet |

#### User Impact

- Helm release deployment fails or times out.
- Deployment does not have desired number of replicas.
- kubectl describe pods shows "ImagePullBackoff"

#### What is shown in Theliv UI?

- '1' issue found in namespace '{{namespace.name}}' inside cluster '{{cluster.name}}'
- Helm release '{{helm.release.name}}' in failed state. ({{helm.release.name}} is a link which shows the helm release details)
- Deployment '{{deployment.name}}' in a failed state ({{deployment.name}} is a link which shows deployment yaml file)
- Pod '{{pod.name}}' has the following errors

Run some default checks.

- Does deployment contain a 'imagePullSecrets' value?
- Does that value exist as a secret?

If NO to any of the above, then

- ###### _possible causes_ (to be shown in UI)
  1. Unable to pull image "docker.xxx.com/sample:v0.0.671" for the container '{{container.name}}'.
     if YES, proceed to further checks.
  2. ImagePullSecret is missing.<br/><br/>
     [OR]<br/><br/>
  3. ImagePullSecret 'regcred' does not exist.

| _if error is following, then show the relevant possible causes_                                                                                                                                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Failed to pull image "docker.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: manifest for docker.xxx.com/appid/sample:v0.0.67 not found: manifest unknown: The named manifest is not known to the registry. |

- ###### _possible causes_ (to be shown in UI)
  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. Either the image repository name 'appid' is incorrect or does NOT exist. <br/><br/>
     [OR]<br/><br/>
  3. Either the image name 'sample' is invalid or does NOT exist.<br/><br/>
     [OR]<br/><br/>
  4. Either the image tag 'v0.0.671' is invalid or does NOT exist.<br/><br/>
     [OR]<br/><br/>
  5. If all the above are correct, then is it probably your ImagePullSecret which is either incorrect or expired.

| _if error is following, then show the relevant possible causes_                                                                                                                                                                             |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Failed to pull image "docker1.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: Get https://docker1.xxx.com/v2/: dial tcp: lookup docker1.xxx.com on xxx.xxx.xx.xx:53: no such host. |

- ###### _possible causes_ (to be shown in UI)
  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. Image registry host 'docker1.xxx.com' is either incorrect or DNS is not able to resolve the hostname.

| _if error is following, then show the relevant possible causes_                                                                                                                                                                            |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Failed to pull image "docker1.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: Get https://docker1.xxx.com/v2/: dial tcp: lookup docker1.xxx.com on xxx.xxx.xx.xx:53: i/o timeout. |

- ###### _possible causes_ (to be shown in UI)
  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. Image registry host 'docker1.xxx.com' is not reachable from kubelet in node abcd xx.xx.xx.xx because of a possible networking issue.
  3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)

| _if error is following, then show the relevant possible causes_                                                                                                                                                                                                                         |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Failed to pull image "docker1.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: Get https://docker1.xxx.com/v2/: dial tcp: lookup docker1.xxx.com on xxx.xxx.xx.xx:53: read udp [::1]:44322->[::1]:53: read: connection refused. |

- ###### _possible causes_ (to be shown in UI)
  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. Image registry host 'docker1.xxx.com' is not reachable from kubelet in node abcd xx.xx.xx.xx because of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues.
  3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)

| _if error is following, then show the relevant possible causes_                                                                                                                                                                                                                         |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Failed to pull image "docker.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: Get https://docker1.xxx.com/v2/: dial tcp: lookup docker1.xxx.com on xxx.xxx.xxx.xx:53: unauthorized or access denied or authentication required. |

- ###### _possible causes_ (to be shown in UI)

  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. ImagePullSecret '{{ImagePullsecret.name}}' is either incorrect or expired.<br/><br/>
     [OR]<br/><br/>
  3. Repository 'sample' does not exist.

| _if error is following, then show the relevant possible causes_                                                                                                                                                                                                                   |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Failed to pull image "docker.xxx.com/appid/sample:v0.0.67": rpc error: code = Unknown desc = Error response from daemon: Get https://docker1.xxx.com/v2/: dial tcp: lookup docker1.xxx.com on xxx.xx.xx.xx:53: Quota execeeded or Too Many Requests or rate limit. |

- ###### _possible causes_ (to be shown in UI)

  1. Unable to pull image "docker.xxx.com/appid/sample:v0.0.671" for the container '{{container.name}}'.
  2. Image registry docker.xxx.com has been ratelimitted. Please increase the quota or limit.

#### Useful links to be shown in the UI

- [kubectl describe command]()
- [view logs in datadog]()
- Button called ['share this issue']() which produces a deep link to this specific issue. Copy the link produced and send it to SRE team member via slack/teamschannel/email who can look at the exact issue just by clicking on the link and help if needed.
- Tell us what you think. Was this helpful? (thumbs up)(thumbs down)

#### Implementation notes

- TODO https://aws.amazon.com/premiumsupport/knowledge-center/eks-ecr-troubleshooting/
- TODO (GKE specific) https://stackoverflow.com/questions/66213804/gke-pulling-images-from-a-private-repository-inside-gcp
- TODO (ICP specific) https://stackoverflow.com/questions/60569646/can-you-pull-docker-images-directly-into-ibm-cloud-kubernetes-clusters
- (no available registry endpoint: failed to do request) https://stackoverflow.com/questions/56515971/microk8s-pulling-image-stuck-in-containercreating-state
- (certificate signed by unknown authority) https://github.com/containerd/containerd/issues/3847
- TODO (GCR specific troubleshooting) https://cloud.google.com/container-registry/docs/troubleshooting

---

## 2.CrashLoopBackOff

| error_code    | status  | tags                     | dependsOn     |
| ------------- | ------- | ------------------------ | ------------- |
| CRASHLOOP_ERR | pending | crashloopbackoff,kubelet | IMAGEPULL_ERR |

#### User Impact

- Helm release deployment fails or times out.
- Deployment may or may not have desired number of replicas (if it is in the middle of the rollout, it will show that current and desired numbers are same).
- Some of the containers in a pod are still not in a Running state.
- kubectl describe pods shows "CrashLoopBackOff"

#### What is shown in Theliv UI?

- '1' issue found in namespace '{{namespace.name}}' inside cluster '{{cluster.name}}'
- Helm release '{{helm.release.name}}' in failed state. ({{helm.release.name}} is a link which shows the helm release details)
- Deployment '{{deployment.name}}' in a failed state ({{deployment.name}} is a link which shows deployment yaml file)
- Container '{{container.name}}' inside Pod '{{pod.name}}' is in a Unhealth state.

| _if error is following, then show the relevant possible causes_                                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Error: failed to start container "tenancymgr": Error response from daemon: OCI runtime create failed: container_linux.go:370: starting container process caused: exec: "exit 120": executable file not found in $PATH: unknown. |

- ###### _possible causes_ (to be shown in UI)

  1. Container '{{container.name}}' has been restarted more than 10 times in the last 120 seconds.
  2. Container '{{container.name}}' has EXITED with a non-zero exit code (127). Check your command or application startup logs.
  3. Give more insights in the UI based on this https://intl.cloud.tencent.com/document/product/457/35758. E.g if exit code is 127, then look at your "command" and make sure it is correct. Problem is there.
  4. Container '{{container.name}}' was unable to start because of the following error.

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Readiness probe failed: HTTP probe failed with statuscode: 404. |

- ###### _possible causes_ (to be shown in UI)

  1. Container '{{container.name}}' has been restarted more than 10 times in the last 120 seconds.
  2. Following readiness probe has failed for the container '{{container.name}}'.

     readinessProbe: <br/>
     failureThreshold: 3 <br/>
     httpGet: <br/>
     path: /readyz <br/>
     port: http <br/>
     scheme: HTTP <br/>
     initialDelaySeconds: 20 <br/>
     periodSeconds: 10 <br/>
     successThreshold: 1 <br/>
     timeoutSeconds: 1 <br/>

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Liveness probe failed: HTTP probe failed with statuscode: 404.  |

- ###### _possible causes_ (to be shown in UI)

  1. Container '{{container.name}}' has been restarted more than 10 times in the last 120 seconds.
  2. Following liveliness probe has failed for the container '{{container.name}}'.

     livenessProbe: <br/>
     failureThreshold: 3 <br/>
     httpGet: <br/>
     path: /healthz1 <br/>
     port: http <br/>
     scheme: HTTP <br/>
     initialDelaySeconds: 20 <br/>
     periodSeconds: 10 <br/>
     successThreshold: 1 <br/>
     timeoutSeconds: 1 <br/>

#### Useful links to be shown in the UI

- [kubectl describe deployment command]()
- [kubectl describe pod command]()
- [kubectl logs command]() kubectl logs <pod-name> --previous
- [view logs in datadog]()

#### Implementation notes

- If "Unhealthy" is the reason, liveliness or readiness probe issue
- If "Failed" is the reason, container startup issue.
- TODO Add a default check to throw a warning if Deployment file does not have a "command?"
- https://medium.com/tailwinds-navigator/kubernetes-tip-how-to-disambiguate-a-pod-crash-to-application-or-to-kubernetes-platform-f6c1395a8d09
- Exit Code 128 & above: Would start debugging from Kubernetes Platform perspective and go up the stack.

---

## 3. RunContainerError, ContainerCreating

RunContainerError— One possible cause: ConfigMap/Secrets missing.
ContainerCreating— Something not available immediately, persistent volume?

TODO

---

## 4. Pending Pods

| error_code       | status  | tags         | dependsOn              |
| ---------------- | ------- | ------------ | ---------------------- |
| PENDING_PODS_ERR | pending | pending pods | CLUSTER_AUTOSCALER_ERR |

Good reference: https://www.datadoghq.com/blog/debug-kubernetes-pending-pods

#### User Impact

1. Helm Release times out.

#### What is shown in Theliv UI?

1. 1 issue found in namespace. You have 1 or more pods in pending state. The pods are [pod1][pod2][..]. When clicked on 'pod1' it can open up a popup which shows the details of "kubectl describe pod pod1 -n namespace"

| _if error is following, then show the relevant possible causes_                                                |
| -------------------------------------------------------------------------------------------------------------- |
| If pod.status.events = FailedScheduling and From = cluster-austoscaler, then jump to cluster autoscaler checks |

if NO cluster autoscaler in the cluster:

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Error: "out of disk space" OR "Insufficient"                    |

- ###### _possible causes_ (to be shown in UI)

1. Add more nodes to your cluster. Your cluster has insufficient capacity.

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Error: "unschedulable"                                          |

- ###### _possible causes_ (to be shown in UI)

1. 1 or more nodes in your cluster is cordoned. Cordoning is an administrative action. Existing pods will continue to run while new pods cannot be scheduled. Please contact your cluster administrator and share this issue with them. (Implementation note: Cordoned nodes report a SchedulingDisabled node status)

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |

| Error: "didn't match node selector" OR "didn't tolerate" |

- ###### _possible causes_ (to be shown in UI)

1. You have nodes in your cluster with specific labels and taints. Your deployment do not match those. You either have to addd the right labels to match the node labels or add tolerations that match the node taints. Read more about node selectors, taints and toleration here -> {{documentation}}

TODO:
PersistentVolume-related scheduling failures
Refer: https://www.datadoghq.com/blog/debug-kubernetes-pending-pods/

#### Useful links to be shown in the UI

- [kubectl describe pod pod1 -n namespace]()

#### Implementation notes

- if pending pods and cluster autoscaler is installed, start the debugging process for cluster autoscaler take into account cluster autoscaler issues if enabled in the cluster.
- Cordoned nodes report a SchedulingDisabled node status. Once it has been cordoned, a node will also display an unschedulable taint (more on taints later) and Unschedulable: true
- Correlate this issue with "NodeNotReady" below
- What to do when cluster autoscaler is in the process of adding new nodes? Should we wait

---

## 5. NodeNotReady or Unkown node status

| error_code        | status  | tags                 | dependsOn |
| ----------------- | ------- | -------------------- | --------- |
| NODE_NOTREADY_ERR | pending | nodenotready,kubelet |           |

Reference: https://stackoverflow.com/questions/47107117/how-to-debug-when-kubernetes-nodes-are-in-not-ready-state
https://aws.amazon.com/premiumsupport/knowledge-center/eks-node-status-ready/

Node status: https://kubernetes.io/docs/concepts/architecture/nodes/#node-status

#### User Impact

- Helm release deployment fails or times out.
- Deployment does not have desired number of replicas.
- Pods might get into pending state because of unavailability of nodes.

#### What is shown in Theliv UI?

- Some of nodes are not stable. Click here for the list.

Default check: See if generic node stats are stable or if node has memory issues etc which could have brought down the kubelet.

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Node.Status.Conditions.Ready = False                            |

###### _possible causes_ (to be shown in UI)

1. Node is not read because of {{Node.Status.Conditions.Ready.Message}}. You can view the relevant kubelet logs by clicking on the link below.
   (If possible detect specific types of error and show the user clear direction e.g. certificate expiry, authentication errors etc)
   (if you detect n/w timeout errors, ask the user to check security group, ability to communicate with API server etc)

| _if error is following, then show the relevant possible causes_                                        |
| ------------------------------------------------------------------------------------------------------ |
| kubelet is down (query for kubelet logs during this timeframe and if entries exist assume it is down ) |

###### _possible causes_ (to be shown in UI)

1. kubelet is either down or not sending logs to the logging system. Make sure kubelet is up and running.

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| If node status is Unknown                                       |

###### _possible causes_ (to be shown in UI)

1. Node status is unknown typically because if the node controller has not heard from the node in the last node-monitor-grace-period (default is 40 seconds). Check either,
   Reference: https://kubernetes.io/docs/concepts/architecture/nodes/#node-status

   1.1 kubelet is down [OR]

   1.2 kubelet is not able to communicate to api server because network is not opened [OR]

   - Confirm that there are no network access control list (ACL) rules on your subnets blocking traffic between the Amazon EKS control plane and your worker nodes. (if aws similar instructions for gcp, azure needs to be printed)
   - Confirm that the security groups for your control plane and nodes comply with minimum inbound and outbound requirements.

   - If your nodes are configured to use a proxy, confirm that the proxy is allowing traffic to the API server endpoints.
   - To verify that the node has access to the API server, run the following netcat command from inside the worker node

     1.3 Amazon EC2 API endpoint is unreachable [if aws]
     Reference: https://aws.amazon.com/premiumsupport/knowledge-center/eks-node-status-ready/

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| If NetworkReady=false or NetworkPluginNotReady                  |

###### _possible causes_ (to be shown in UI)

1. kubelet is not ready because CNI drive is not ready. Please check whethere the CNI driver is running fine on the node {{node.name}}.
   You can find the kubelet logs and relevant CNI logs by clicking on the link below.

#### Useful links to be shown in the UI

- List of nodes NotReady [node1] [node2] [..] . When clicked on 'node1' it can open up a popup which shows the details of "kubectl describe node node1"
- [kubelet logs deeplink] when clicked on takes you to the centralized logging system with right filters like timeframe etc
- [cluster autoscaler logs deeplink]
- [kubectl logs clusterautoscaler](entire kubectl logs command which can be copied to clipboard)
- [kubectl logs cni-drive/aws-node/weavenet/flannel](entire kubectl logs command which can be copied to clipboard)

#### Implementation notes

- What to do when cluster autoscaler is in the process of adding new nodes? Should we wait
- Use Cluster-Level Events
  K8S fires events whenever the state of the resources it manages changes (Normal, Warning, etc). They help us understand what happened behind the scenes. The get events command provides an aggregate perspective of events.
  e.g
  all events sorted by time kubectl get events --sort-by=.metadata.creationTimestamp
  warnings only kubectl get events --field-selector type=Warning
  events related to nodes kubectl get events --field-selector involvedObject.kind=Node
  Reference: https://betterprogramming.pub/5-easy-tips-for-troubleshooting-your-kubernetes-pods-34f594e03ba6

# all events sorted by time.

kubectl get events --sort-by=.metadata.creationTimestamp

# warnings only

kubectl get events --field-selector type=Warning

# events related to Nodes

## kubectl get events --field-selector involvedObject.kind=Node

## 6. ClusterAutoscaler event debugging

| error_code             | status  | tags                          |
| ---------------------- | ------- | ----------------------------- |
| CLUSTER_AUTOSCALER_ERR | pending | clusterautoscaler,autoscaling |

https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-events-are-emitted-by-ca

https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-autoscaler-visibility#debugging_scenarios

Check if cluster autoscaler is up and running. In version 0.5 and later, it periodically publishes the kube-system/cluster-autoscaler-status config map. Check last update time annotation. It should be no more than 3 min (usually 10 sec old).

Check in the above config map if cluster and node groups are in the healthy state. If not, check if there are unready nodes.

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Detected events like noScaleUp or noScaleDown                   |

Refer: https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-autoscaler-visibility#noscaleup-top-level-reasons (Is this GKE specific?)

Refer: https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md#what-events-are-emitted-by-ca (some failure events are NOT recorded in status config map, rather they are present only on nodes/pods e.g. ScaleDownFailed event)

###### _possible causes_ (to be shown in UI)

1. Cluster autoscaler addons is unstable. It is not able to add more nodes because of .. [OR]
1. Cluster autoscaler addons is unstable. It is not able to delete 1 or more nodes becaus of
   Contact your cluster administrator or whoever is responsible for maintaining cluster autoscaler.

#### Implementation notes

- Take this into account: Cluster Autoscaler also doesn't trigger scale-up if an unschedulable pod is already waiting for a lower priority pod preemption.

- What to do when cluster autoscaler is in the process of adding new nodes? Should we wait

---

## 7. Management namespaces instability

| error_code                   | status  | tags                  |
| ---------------------------- | ------- | --------------------- |
| MGMT_NAMESPACES_INSTABLE_ERR | pending | management namespaces |

This refers to 9.0.0 [here](features.md)

#### User Impact

- Anything ranging from deployments timings to ingress not working all of a sudden.

#### What is shown in Theliv UI?

- One or more management namespaces are unstable. Management namespaces typically hold the common or cluster level addons which is required to be in stable for the cluster to function properly. Contact your cluster administrator for this issue.
- [kube-system] [core-addons][...] are unstable (clicking on kube-system can describe the namespace)
- Following addons are unstable

  - [deployment1]
    - [pod1] is in crashloopbackoff [pod event here]. This is because [clear explanation]. You can view logs using this kubectl command here [here](kubectl logs ...). You can view the logs in datadog [here] (datadog deep link)

- ###### _possible causes_ (to be shown in UI)

1.  Same as pods crashloop back off, pending, terminating pod states

#### Useful links to be shown in the UI

- The links can be shown inline instead of a separate pane for this usecase.

#### Implementation notes

---

## 8. Pods stuck in Terminating or Unknown state

https://kubernetes.io/docs/concepts/architecture/nodes/#node-status
The node controller does not force delete pods until it is confirmed that they have stopped running in the cluster. You can see the pods that might be running on an unreachable node as being in the Terminating or Unknown state.

TODO

---

## 9. Deployment/Service definition discrepancies

Refer:

- deployment port and service port mismatch https://stackoverflow.com/questions/67472659/unable-to-connect-to-the-api-deployed-in-docker-desktop-kubernetes-loadbalance
  https://stackoverflow.com/questions/67385762/getting-502-bad-gateway-nginx-while-accessing-java-application-deployed-on-kub

- service and deployment does not have matching labels https://stackoverflow.com/questions/67387937/error-when-exposing-kubernetes-service-through-nginx-ingress-and-file-type-load

TODO

---

Pending TODO

1. Detecting Resource quota/Limit range related issues user face. Any recent evictions/terminations because of this? https://www.datadoghq.com/blog/monitoring-kubernetes-performance-metrics/
2. Support for kubernetes pod status ERROR instead of Crashloopbackoff
3. Detecting Node eviction, OOM killed pods issue. Is it worth including this? Should these be warnings? (kube_pod_container_status_terminated_reason ) https://github.com/kubernetes/kube-state-metrics/blob/master/docs/pod-metrics.md
   1. Refer https://sysdig.com/blog/troubleshoot-kubernetes-oom/
   2. https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/
   3. https://medium.com/tailwinds-navigator/kubernetes-tip-how-does-oomkilled-work-ba71b135993b
   4. https://sysdig.com/blog/kubernetes-pod-evicted/
   5. sum_over_time(kube_pod_container_status_terminated_reason{reason!="Completed"}[5m]) > 0 to detect a recent non-graceful termination (https://github.com/kubernetes/kubernetes/issues/69676#issuecomment-439718989)
   6. https://github.com/kubernetes/kubernetes/pull/87856
      Template:
4. Is node allocatable properly configured? Is kubelet starting to reclaim resources based on node's memory pressure etc. (root cause= node memory pressure, symptom=pod evictions).
5. https://www.datadoghq.com/blog/monitoring-kubernetes-performance-metrics/ has a lot of scenarios
6. https://kubernetes.io/blog/2021/04/13/kube-state-metrics-v-2-0/
   1. kube_pod_container_status_restarts_total can be used to alert on a crashing pod.
   2. kube_deployment_status_replicas which together with kube_deployment_status_replicas_available can be used to alert on whether a deployment is rolled out successfully or stuck.
   3. kube_pod_container_resource_requests and kube_pod_container_resource_limits can be used in capacity planning dashboards.
7. https://epsagon.com/development/how-to-guide-debugging-a-kubernetes-application/

#### User Impact

#### What is shown in Theliv UI?

| _if error is following, then show the relevant possible causes_ |
| --------------------------------------------------------------- |
| Error:                                                          |

- ###### _possible causes_ (to be shown in UI)

#### Useful links to be shown in the UI

#### Implementation notes
