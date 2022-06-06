# Developer Guide
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
Check the valid [metrics provided by kube_state_metrics](https://github.com/kubernetes/kube-state-metrics/tree/master/docs) before defining your own alerts expressions.
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
