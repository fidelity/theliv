### Implementation difficulties/Things to consider

- Explore the idea of "warnings" category. This could be based on alerts (e.g impending issues but not a blocker. e.g alert says pod regularly exceeds the set the request which means the request needs to be increased)
- Explore how we do exploit and make use of https://kubernetes.io/docs/concepts/workloads/pods/ephemeral-containers/

- How do we make this system maintenable? i.e if we support automatically detecting "ImagePullBackOff" issue and if one of the kubernetes release change the logic that is required to detect it, how do detect it. How should the e2e testing be structured so that such issues can be detected. How complex can this get?

- Should the debugging process take into account hierarchy e.g should crashloop back off be debugged after it debugs imagepullbackoff or should we treat each issues as totally independent. Could be an important detail that drives the implementation logic.

- [feature] Agent based system? Should we run an agent or rely on datadog to analyze kubernetes events as "starting" point to figure out more about a specific problem? E.g when systems STARTS the analysis, if it detects kind deployment not in a good state, it gets a start time of the deployment (when issue started) and current time (or end time where issue occurred) and between those times, it queries events to see if there was any other Error event that occurred that could have affected this? What if there was Node Not ready event during that time? This needs lots of thinking.

- [feature] Should we get a custom timeframe from the user? e.g last 1 hour? which can give theliv an approximate timeframe to check if something went wrong during that time?

- [feature] each issue will have a unique ID or error code e.g KUBE_CRASHLOOP_ERR? and some attached metadata tags. for such ID, configure a setup a KNOWLEDGE answers, optional confluence page.

- highly complex issue of issue vs version compatibility matrix? what if an issue is specific one of kubernetes release? how do we differentiate it. Should be add tags like <= 1.20.x etc? How complex is this scenario

- As a part of ### 11.0.0 Native support for commonly used addons, do we need to get the location of control plane logs. Do we need to use/analyze scheduler logs etc? Is it worth providing deep link for these logs?

- Do we need to indicate warning for known security vulnerabilities etc if we have internet access from theliv?

- Do we need to take this into account https://docs.aws.amazon.com/eks/latest/userguide/troubleshooting.html

- Should we ignore/abort if a deployment is in progress or wait till it is completed and recalculate?

- Along with showing errors/issues should there be warnings? These could be "informational" or best practices (which can be turned off by the admininistrator who sets it up). E.g. setting up resources quotas/limit ranges for pods, possible vulnerable release etc.

- Should we detect API deprecations and report as error or warning? https://kubernetes.io/docs/reference/using-api/deprecation-guide/

- How do we debug CRDs? Can we get the user from input which tell theliv how to detect failure in CRDs? e.g. in the YAMl specifi Status.Phase != "Suceeded", >5m should be considered as failure. We can try supporting wellknown CRDS like kafka, elastic search etc?

- Should we use auditlogs, control plane logs (controller manager etc) to detect some issues? How feasible is this?

- If a calico n/w policy block a traffic, should we analyse the events emitted related to this and report potential pod -> pod communication issues? Should we include VPC flow logs and include any potential denials happening in outbound traffic to node?

- [feature] Can we have dedicated set of alerts managed exclusively by theliv against well known telmetry tools which i can listen to and use it while debugging? e.g. have i received any alerts for this namespace from splunk or datadog for an exiting deployment?
- Analysis based on API server metrics exposed by the control plane
