# Developer Guide
This guide includes 2 parts: How to customize your own alerts and how to register investigator for the alert.
## Customize Alerts
The alerting rules are defined in the config map of prometheus server. You can see all the configurations with the following command.
``` cmd
kubectl get cm prometheus-server -o yaml
```
You should see alerting_rules.yml in the config map, under which part are the alerts placed.
``` cmd
apiVersion: v1
data:
  alerting_rules.yml: |
    groups:
    - name: kube-state-metrics
      rules:
      - alert: <YourAlertHere>
  ...
```
### Define Alert
Theliv mainly uses **kube-state-metrics** to define prometheus alerts, which is a simple service that listens to the Kubernetes API server and **generates metrics about the state of the objects such as deployments, nodes, and pods.** The metrics are focused on orchestration metadata: deployment, pod, replica status, etc. It is important to note that [kube-state-metrics](https://github.com/kubernetes/kube-state-metrics) is just a metrics endpoint. Other entities need to scrape it and provide long term storage (e.g., the Prometheus server).
#### Create Alert
Check the valid [metrics provided by kube_state_metrics](https://github.com/kubernetes/kube-state-metrics/tree/master/docs) before defining your own alerts expressions. And specify the label *job='kube-state-metrics'* in the metric.  
[Prometheus alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) allow you to define alert conditions based on Prometheus expression language expressions and to send notifications about firing alerts to an external service.
**Alert Example:**
``` cmd
      - alert: PodNotRunning
        expr: kube_pod_status_phase{job='kube-state-metrics',phase=~'Failed|Pending|Unknown'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0
        for: 5m
        labels:
          resourcetype: pod
          # The resourcetype label is REQUIRED for Theliv. 
          # You should specify the kind of resource (container, pod, node, deployment etc.) that triggers the alert.
        annotations:
          summary: "Pod is not in Running status"
          description: "Pod {{$labels.pod}} in namespace {{$labels.namespace}} is in {{$labels.phase}} status for more than 5mins."
```
**Note:** The label ***resourcetype*** is required in Theliv. You should specify the kind of resource (container, pod, node, deployment etc.) that triggers the alert.
#### Update Configuration
Edit the prometheus configuration to add your alerts:
``` cmd
kubectl edit cm prometheus-server
```
#### Reload Configuration
Reload prometheus configuration.
#### References
[kube_state_metrics docs](https://github.com/kubernetes/kube-state-metrics/tree/master/docs)   
[PromQL Cheetsheet](https://promlabs.com/promql-cheat-sheet/)   
[Alerts Examples](https://github.com/kubernetes-monitoring/kubernetes-mixin/blob/c76b9378b86d28bd617d94a57c72b4770efed510/alerts/apps_alerts.libsonnet)   

## Provide Investigator
Investigators are mainly responsible to triage the problem that built from a fired alert and provide responding solutions and enough details for users to understand the next steps to solve the problem.  
In this example, we have an alert to monitor the init container's ImagePullBackoff error.
``` cmd
      - alert: InitContainerWaitingAsImagePullBackOff
        expr: kube_pod_init_container_status_waiting_reason{job='kube-state-metrics', reason=~'ImagePullBackOff|ErrImagePull|InvalidImageName'} * on(uid, container) group_left(image) kube_pod_init_container_info{job='kube-state-metrics'} * on(uid) group_left(owner_kind, owner_is_controller, owner_name) kube_pod_owner{job='kube-state-metrics'} >0
        for: 5m
        labels:
          resourcetype: container
          # resourcetype: initcontainer
          # either initcontainer, or container is okay
        annotations:
          summary: "Init Container is Waiting due to ImagePullBackOff"
          description: "Init Container {{$labels.pod}} of Pod {{$labels.pod}} in namespace {{$labels.namespace}} is waiting due to ImagePullBackOff, reason is {{$labels.reason}}"
```
And we need to provide the responding investigator:
1. If the alert is to monitor a new resource type that not exist in *detector.go -> buildProblemAffectedResource*:
   1. You need to modify the func *buildProblemAffectedResource* to load its runtime object.
   2. You need to specity the problem level. The problem with the lowest level will become the root cause. (container level is lower than deployment level, thus container failure is the root cause of a deployment failure)
2. Create a new investigator file under */internal/investigators* for the resource type, in this example is *initcontainerinvestigator.go*
3. Create an investigator function *InitContainerImagePullBackoffInvestigator* in the file.
4. Register the function in the *alertInvestigatorMap* in */pkg/service/detector.go*. The key is the with alert name *InitContainerWaitingAsImagePullBackOff*. The value is one or more investigator functions you expect to execute for the alert.
5. Implement the investigator function and build *problem.SolutionDetails*. In this example we use go template to provide solutions formatting.
6. After above steps, you should see *issue.solutions* in response.
```json
[
  {
    "name": "problem name",
    "rootCause": { ... },
    "resources": [
      {
        "name": "pod-name-with-error",
        "type": "Pod",
        "labels": { ... },
        "annotations": { ... },
        "metadata": { ... },
        "issue": {
          "name": "InitContainerWaitingAsImagePullBackOff",
          "description": "Description of the alert",
          "solutions": [
            "!!!!! Here are the solutions you expect to see !!!!!",
            "1. Unable to pull image xxxx.com/nginx:notexist for the container nginx. The root cause could be one of the following.",
            "2. Either the image repository name is incorrect or does NOT exist.",
            "3. < You can also provide enough details to troubleshoot >",
            "4. Either the image tag is invalid or does NOT exist.",
            "5. < You can also provide cmd to troubleshoot >"
          ]
        },
        "tags": { ... },
        "causelevel": 1
      }
    ]
  }
]
```
[Example Investigator](https://github.com/fidelity/theliv/blob/sample-investigator/internal/investigators/initcontainerinvestigator.go)