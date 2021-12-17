# Feature backlog

## 1.0.0 Integrate with deployment pipelines by providing deep links

| status  | tags                                     |
| ------- | ---------------------------------------- |
| pending | medium-priority,pipeline, cicd, deeplink |

_Deep link is a typically a lengthy url with lots of query parameters which adds lot of context when the URL is processed._
Whenever a deployment job fails in any deployment pipeline (jenkins, uDeploy, harness etc), the deployment pipeline will print out a deeplink URL for theliv at the end of the logs.
e.g https://theliv.org.com/v1/kubernetes?namespace=dev-ns&cluster=dev-cluster-1&cloudprovider=aws&platform=eks
Whenever user sees a deployment job fail, instead of scrolling through a vast amount of logs, they can simply look for a theliv link at the bottom of the logs and click on it which will open up a UI and tell the user exact what has gone wrong in that namespace in that cluster.
So, provide an API endpoint which the deployment pipelines can call to retrieve this deeplink URL that can be printed out at the end of logs. When the deployment pipeline calls the API, it sends the relevant context it already has i.e namespace, cluster, cloud provider, platform details etc and gets back a proper deeplink URL.

Today for example, users scroll through the vast set of logs to figure out the error printed e.g "Helm release timed out" and then they gather some basic data like cluster, namespace details etc, generate a kubeconfig, connect to the cluster and run a various set of commands to arrive at the issue. While some developers are comfortable with this, vast majority of them either dont know or dislike this debugging process.

With this feature, theliv enhances the user experience significantly by allowing them to simply click on a link which opens up a dashboard that automatically analyses the ongoing issues inside the namespace and reports it to the user in a detailed way.

## 2.0.0 Recording each finding into a database

| status  | tags             |
| ------- | ---------------- |
| pending | high-priority,db |

Before the issue is reported the user, it needs to be saved into a kv database (or a document db?) using a unique id. Using this unique id, the issue found can then be easily "shared" to someone else.

## 3.0.0 'Share this issue' button

| status  | tags          |
| ------- | ------------- |
| pending | high-priority |

Today many of the technical forums (internal or external ones like stackoverflow) follow a model where the user asks a question by explaining the issue in detail and someone jumps into help them out. In the process of helping with that issue, the user might be asked for more details via conversation in the comments section.

Because theliv finds the issue, it has all the context around that issue. E.g which namespace & cluster this happened, when did this happen. Can we see more/relevant logs in Splunk/ELK/Datadog? etc. All the stuff that a user typically needs to debug an issue can be found here. Hence if the app user needs further help, e.g he wants to get more help from their teammate or an devops team member, they should be able to simply click on a "share this issue" button which generates a link with the unique which can then be shared with the other user.

So instead of the user explaining the issue in detail in a forum, they can simply provide this link in the interal forums and ask for more help. This way, the user who is going to help the app team member (it could be devops or an SRE team member) has all the context they need about the issue in one place.

## 4.0.0 Helpful documentation links for the issues found

| status  | tags              |
| ------- | ----------------- |
| pending | low-priority, TBD |

Every issue type (e.g ImagePullBackoff is an issue type. CrashloopBackoff is another one) will have an unique id or an error code and some default documentation mapped to it. It could be kubernetes documentation which explains more about the details surrounding that issue. Optionally we need to give a way for the administrator or a platform team inside an organization to plug in their custom documentation (confluence link etc). This can be got as in put via a configuration yaml file etc where the administrator or platform team in an organization can map these unique error codes to confluence links. This way when theliv reports its finding for a particular issue, towards the end it can display kubernetes documentation that might help the user around that issue AS WELL AS the custom confluence links which could be organization specific.
_Note: There is a maintenance problem here where the links could be changed or moved. We will rely on PRs from users who can report the broken links etc until we find a better way_

## 5.0.0 Pluggable logging driver

| status  | tags            |
| ------- | --------------- |
| pending | medium-priority |

Almost all the organization will have a centralized logging system where all the logs are sent. These are infrastructure logs, event logs, application logs etc. Theliv should support plugging in various logging drivers like Splunk, ELK, Datadog etc. It should start supporting one by one. The reason for this is when theliv finds the issue and reports it in the UI, it will also generate relevant deeplink URLs for the logging driver which can be used the user to check relevant logs if necessary. Without this, user typically logging into the logging system's UI and manually form the filters (e.g start and end time, namespace, cluster, index ids etc and many other logging driver specific filters). Lot of enterprise organizations are very sensitive about who can see the logs etc and they rely on logging system's RBAC model to control this access. Because of this theliv UI should not show the logs rather will simply form the deeplink URLs which can be clicked on by the user to look at the details logs after authenticating to their centralized logging systems.

_**Note:** Shoudl be extended to include tracing driver, metrics driver etc in future._

## 6.0.0 Solid e2e testing

| status  | tags          |
| ------- | ------------- |
| pending | high-priority |

One of the obvious problems with theliv is the maintenability of the issues it finds. What if the logic to debug "ImagePullBackOff" needs to be changed because of a new kubernetes release? How do we catch those? This where we need to have a solid e2e testing framework which reproduces the specific issue and runs theliv against the sytem to make sure it captures it properly. This is easily one of the important pillars behind theliv.

## 7.0.0 RBAC

| status  | tags         |
| ------- | ------------ |
| pending | low-priority |

## 8.0.0 "Re-run" a specific check

| status  | tags         |
| ------- | ------------ |
| pending | low-priority |

If an issue is found and the relevant link is "shared" with another user and if the user happens to check if after sometime, he might like to "re-run" or "re-check" if the issue STILL persists. So an option to re-run a check would be helpful which can retrive the latest status.

## 9.0.0 Support concept of mgmt namespaces or namespace dependencies

| status  | tags          |
| ------- | ------------- |
| pending | high-priority |

It is a common practice now to deploy core addons or common addons (ingress, telemetry collector, opa etc) into a dedicated set of namespace which are typically referred to as management namespaces (kube-system being one of default mgmt namespace). So if any of the addons in these namespaces are unstable, then report them and no need to check further. These management namespaces are supposed to be stable and up and running all the time.

## 9.1.0 Customizable logic to check namespace dependencies

| status  | tags         | dependencies |
| ------- | ------------ | ------------ |
| pending | low-priority | 9.0.0        |

9.0.0 talks about management namespaces being stable. The definition of stable can become opinionated so it might be better to expose some pluggable framework. e.g base framwork will check if pods are running and healthy but a custom logic that needs to be plugged into can additionally check for any ongoing alerts for those namespaces (based on errors/exceptions addons logs etc). These alerts might be fired even when the pods are in "running" state but can be unhealthy.

## 10.0.0 Easy way for user to give feedback

| status  | tags         |
| ------- | ------------ |
| pending | low-priority |

What if a user sees this and thinks theliv is not giving out right information (in terms of possible causes, next steps etc) or what is it is giving out outdated information. It might be helpful to have link next "click to report a problem" which allows user to raise a github issue easily (prefilled with some information about the problem, relevant tags so that right developers get notified etc).

## 11.0.0 Native support for commonly used addons

| status  | tags            |
| ------- | --------------- |
| pending | medium-priority |

Every kubernetes cluster usually has a list of core addons. This could be ingress controller, external-dns, metrics server, CNI driver, cluster autoscaler etc. Sometimes debugging process might need to take these into account. E.g. Nodes are not ready or having issues, might have something to do with cluster autoscaler. So generating "deep links" for cluster autoscaler logs in this case and showing in the UI might be beneficial. Where the user can simply click on the link to see the relevant logs. Sometime nodes not being ready could be something to do with kubelet (though not an addon) and hence kubelet logs might need to be looked into. So having a deep link right there for your logging system could be very handy. That link when clicked on could show the specific set of logs i.e kubelet logs for a specific node, specific cluster, specific timeframe etc.

So, as a one time setup, initially get a config from the user (yaml is fine) where he can configure the core addons that are running in the cluster. This could be a specific list to choose from e.g. CNI driver we can support aws-cni, azure-cli, flannel etc. They can also specify the name or tag that is typically used in the centralized logging system to query.

This information about addons will be useful in cases where for example, if pods are pending, all nodes are in Ready state, check if cluster autoscaler is down (because if it was up and runnign without issues it should have responded to pending pods). Such correlation become easier with this feature where user specifies what core addons they are using.

To begin with we can start supporting the following

- cni driver
- external-dns
- metrics server
- kubelet
- cluster autoscaler

## 12.0.0 Support common root checks

| status  | tags          |
| ------- | ------------- |
| pending | high-priority |

'root checks' are the core checks that needs to happen first. Proceed to debug kubernetes specific issues only if these checks passed. we can start with three to begin.

- global_checkpoint
- cluster_highlevel_checkpoint
- mgmt_namespaces_checkpoint

These can run everytime. We can optimize it later to re-use some of recently run ones (if the check was run in the last min, we dont have to run it again, but this can be futuristic).

_global_checkpoint:_ Check for general infrastructure health. Are there any active incidents from AWS (region failure, az failure etc). Generic network outages, is there any issue related to Direct connect, TransitGateway etc.

_cluster_highlevel_checkpoint:_ Are nodes in ready state? Are the kubelets healthy? Are the API server accessible and fine? Any API server alerts going on? Any API server metrics indicating potential slowless/issue etc?

_mgmt_namespaces_checkpoint:_ Are the management namespaces healthy. i.e Are all the pods running as expected without any crashloop back off etc. Anything in pending state.

## 12.1.0 Extensible common root checks

| status  | tags            |
| ------- | --------------- |
| pending | medium-priority |

The root checks mentioned in 12.0.0 should be extensible. E.g An network team in an organization might configure some fatal alerts in datadog, so a custom check would call to check if any such alerts are fired recently (or ongoing) and take that into account to calculate the global health.

## 13.0.0 Display a flowchart or any type of diagram explaining high level debugging steps carried out

| status  | tags         |
| ------- | ------------ |
| pending | low-priority |

It might be helpful to explain the high level steps or flow chart or tree diagram that was carried out during the debugging process. This can be built into the base framwork. It could be tab next to the actual result page where if a user clicks on it, they will see on high level global_checkpoint, cluster_highlevel_checkpoint etc passed before arriving at the issue. Could be low priority item.

## 14.0.0 Recommended next Steps

| status  | tags          |
| ------- | ------------- |
| pending | high-priority |

Every report should have "Recommended next steps" which will point to the next steps for the user which can include the documentation.

## 15.0.0 Support filters within namespace

| status  | tags         |
| ------- | ------------ |
| pending | low-priority |

By default, theliv checks the health of all possible resources inside a namespace. But sometimes users might want to specific filter (helmrelease: <>, deployment:<> etc). Support such filters where theliv checks health of just those resources.
