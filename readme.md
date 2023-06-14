# Theliv

###### `Don't spend time, debugging the obvious!`

<br><br>
Theliv (pronounced as **_thae-lee-v_**) is a system that aims to help development teams, SRE, DevOps team members by debugging many of the well known infrastructure issues (k8s to begin with) on their behalf and generate a detailed report which details the specifics of the issue, possible root causes and clear next steps for the user facing the problem. In short, instead of you having to open up terminal, run several commands to arrive at an issue, a system does it for you and shows it in a neat UI. That is theliv.

_If you Google "Kubernetes troubleshooting guide" today, you will see thousands of articles, blog posts detailing how to debug common issues in kubernetes. Think of theliv as a framework "codifying" all those troubleshooting guides, where it intelligently debugs the system to find if any issue exists in the system. Now imagine if this can be done not just for kubernetes but for many other platforms out there like kafka, elastic search, RDS etc_

Any technical forum has questions that can roughly be placed into two buckets.

1. knowledge based questions i.e starting with how to.., what is.. etc.
2. Issue based i.e why am I seeing this error/exception etc.

(1) above is typically solved by building another platform abstraction on top of kubernetes, automating as much of possible. Theliv tackles (2)nd problem. Especially the well known & recurring issues. Theliv will detect such issues and provide solutions and next steps to the users. This will significantly reduce L1, L2 teams' workload where they are relieved from repetitively debugging well known issues. While we will start with supporting Kubernetes, theliv aims to add support for more and more infrastructure services in the future where it can detect well known issues with various AWS, Azure and GCP service for example and provide solutions, clear next steps for the user. Clear next steps is the key here.

Theliv also handles correlations. Many of times a problem arises because of cascading effect i.e A leads to B leads to C and usually 'C' is what is seen by the user. Theliv aims to handle this correlation as much as possible where it can notify the users that the 'actual problem' is somewhere else.

Because theliv collects all the related information w.r.t a problem it detects, it becomes easy for you to request help by simply sharing the issue link with some else. This is in contrast to how you have to share every bit of information verbally in a forum while seeking help. With theliv, you just share a link and seek help if you need to.

Let's say you are a developer using your devops teams' deployment pipeline in Jenkins and your deployment job fails. You will have to scroll through the jenkins logs, you find "Helm Release timed out". You then have to run some kubectl commands to see the pods are pending. The pending status can be of multiple reasons, you will have to debug further to figure out which of these reasons is that. If the reason was no nodes available, then you need to debug further to understand why no nodes are available? why is cluster autoscaler not doing its job? you need to login to splunk/datadog to look at several logs to understand more. This can get crazy but to be honest, in the kubernetes world, this is more like a "moderate" issue, you will see lot more complex issues. While a developer can perfectly be equipped with the necessary skills to debug using the above flow, most of the time they don't want to do it since their focus is on rolling out a business feature to production asap. They are in a different kind of pressure to deliver a business logic and that never accounts for the time and effort they put into debugging infrastructure issues. They usually are forced to reach out or push the debugging part of SRE/DevOps team members. The SRE/DevOps team member over a period of time are usually overwhelmed by debugging these "relative repetitive" issues again & again. Over a period of time, they come to think development teams must have the kubernetes knowledge to debug such issues on their own. So, this develops a disconnect between SRE/DevOps teams and development teams over a period of time eventually hampering the cloud native adoption journey of the development teams.

So, in a way a system like Theliv is absolutely necessary to 'truly' scale cloud native architecture to thousands of developers in an organization.

Finally, theliv aims to be an extensible framework where custom checks can be plugged in whereever required.

### Goals

1. Make it extremely easy for developments teams to figure out whether an issue happening is a application issue or an infrastructure issue. It is quite normal for developers to "suspect" underlying infrastructure issue if something starts to fail after a system/platform upgrade. Theliv should clarify this.

2. Correlate multiple issues to see if one could have affected another. Show the root of the problem rather than just the symptoms which user typically sees. E.g Multiple factors could come into play in a specific issue. What if your pods are crashing because your CNI driver is not able to retrieve free IPs? What if your pods are in 'pending' status because cluster autoscaler is having issues? What if your ingress is not setup properly because of a new release of ingress controller rolled into your cluster?

3. Theliv takes of care of showing most of the issues to the development teams where they will further reach out to SRE/DevOps team members only if it is absolutely required.

4. Get a minimal input from the user and theliv figures out the rest. e.g. cluster-name, namespace details are received as input from the user and theliv tries to identify all possible problems that could happen within that namespace and shows if it finds any.

### Non-Goals

Theliv should NOT be another k8s dashboard. There are lots of dashboards out there today that already does it. E.g K8s dashboard, vmware tanzu octant (https://github.com/vmware-tanzu/octant), Lens IDE etc. While such tools are still very useful, it still depends on how good the developer is comfortable with debugging an issue using those tools. Theliv in contrast does not project data using which you can debug, rather debugs the problem on your behalf and shows it to you. Large enterprises with tens of thousands on developers with varying knowledge on infrastructure in general, will still need some tool that will tell them directly what an issue is. Theliv aims to be that tool.

## Contributions

Contributions are very welcome. Please read the [contributing guide](CONTRIBUTING.md) or see the docs.
