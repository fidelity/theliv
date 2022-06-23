/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"strings"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"
)

const (
	UnknownManifestMsg    = "The named manifest is not known"
	NoSuchHostMsg         = "No such host"
	IOTimeoutMsg          = "I/O timeout"
	ConnectionRefused     = "Connection refused"
	UnauthronizedMsg      = "Unauthorized or access denied or authentication required"
	QuotaRateLimitMsg     = "Quota exceeded or Too Many Requests or rate limit"
	RepositoryNotExistMsg = "Repository does not exist or may require 'docker login'"
)

const (
	UnknownManifestSolution = `
1. Unable to pull image  {{ .Image }} for the container {{ .Name}}. The root cause could be one of the following.
2. Either the image repository name is incorrect or does NOT exist.
3. Either the image name is invalid or does NOT exist.
4. Either the image tag is invalid or does NOT exist.
`
	NoSuchHostSolution = `
1. Unable to pull image {{ .Image }} for the container {{ .Name}. The root cause could be one of the following.
2. Image registry host is either incorrect or DNS is not able to resolve the hostname.
`
	IOTimeoutSolution = `
1. Unable to pull image {{ .Image }} for the container {{ .Name}. The root cause could be one of the following.
2. Image registry host is not reachable from kubelet in node because of a possible networking issue.
3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)
`
	ConnectionRefusedSolution = `
1. Unable to pull image {{ .Image }} for the container {{ .Name}. The root cause could be one of the following.
2. Image registry host is not reachable from kubelet in node because of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues.
`
	UnauthronizedSolution = `
1. Unable to pull image {{ .Image }} for the container {{ .Name}. The root cause could be one of the following.
2. ImagePullSecret is either incorrect or expired.
3. Repository does not exist.
`
	QuotaRateLimitSolution = `
1. Unable to pull image {{ .Image }} for the container {{ .Name}. The root cause could be one of the following.
2. Image registry has been ratelimitted. Please increase the quota or limit.
`
)

const (
	SecretMsg1 = `Run the following command to make sure your imagepull secret is correct. Make sure under 'auths', a entry corresponding to you registry hostname exists with the CORRECT username & password. Sometime incorrect imagepull secret might lead to this error as well.
kubectl get secret {{ range .Spec.ImagePullSecrets }}{{ .Name }}{{ end }} -n {{ .ObjectMeta.Namespace }} --output="jsonpath={.data.\.dockerconfigjson}" | base64 --decode
`
	SecretMsg2NotExist = `We see that pod does not reference to any imagePullSecret. Kindly make sure this was intentional and not missed by mistake. Imagepullsecrets are mandatory in order to pull image from a registry that requires authentication and does not support anonymous pulls
`
	SecretMsg3 = `Run the following command to get the imagePullSecret name:
kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} --output="jsonpath={.spec.imagePullSecrets}"
`
)

var ImagePullBackoffSolutions = map[string]func(ctn *v1.Container, msg *string) []string{
	UnknownManifestMsg:    getImagePullBackoffSolution(UnknownManifestSolution),
	RepositoryNotExistMsg: getImagePullBackoffSolution(UnknownManifestSolution),
	NoSuchHostMsg:         getImagePullBackoffSolution(NoSuchHostSolution),
	IOTimeoutMsg:          getImagePullBackoffSolution(IOTimeoutSolution),
	ConnectionRefused:     getImagePullBackoffSolution(ConnectionRefusedSolution),
	UnauthronizedMsg:      getImagePullBackoffSolution(QuotaRateLimitSolution),
}

func ContainerImagePullBackoffInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	pod := *problem.AffectedResources.Resource.(*v1.Pod)
	for _, ctnstatus := range pod.Status.ContainerStatuses {
		detail := ctnstatus.State.Waiting.Message
		var solutions, secretmsg []string
		for _, ctn := range pod.Spec.Containers {
			if ctnstatus.Name == ctn.Name {
				for msg := range ImagePullBackoffSolutions {
					if strings.Contains(strings.ToLower(ctnstatus.State.Waiting.Message), strings.ToLower(msg)) {
						solutions = ImagePullBackoffSolutions[msg](&ctn, &msg)
						secretmsg = checksecretmsg(msg, pod)
					}
				}
				if len(solutions) == 0 {
					msg := UnknownManifestMsg
					solutions = ImagePullBackoffSolutions[UnknownManifestMsg](&ctn, &msg)
					secretmsg = checksecretmsg(msg, pod)
				}
			}
		}
		problem.SolutionDetails = append(problem.SolutionDetails, detail)
		problem.SolutionDetails = append(problem.SolutionDetails, solutions...)
		problem.SolutionDetails = append(problem.SolutionDetails, secretmsg...)
	}
}

func ContainerCrashLoopBackoffInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	// TODO
}

func checksecretmsg(msg string, pod v1.Pod) []string {
	var secretmsg []string
	if msg == UnknownManifestMsg || msg == RepositoryNotExistMsg {
		if len(pod.Spec.ImagePullSecrets) == 0 {
			s := "5. " + SecretMsg2NotExist
			secretmsg = GetSolutionsByTemplate(s, &pod, true)
		} else {
			s := "5. " + SecretMsg1 + "6. " + SecretMsg3
			secretmsg = GetSolutionsByTemplate(s, &pod, true)
		}
	} else if msg == UnauthronizedSolution {
		if len(pod.Spec.ImagePullSecrets) == 0 {
			s := "4. " + SecretMsg2NotExist
			secretmsg = GetSolutionsByTemplate(s, &pod, true)
		} else {
			s := "4. " + SecretMsg1 + "5. " + SecretMsg3
			secretmsg = GetSolutionsByTemplate(s, &pod, true)
		}
	}
	return secretmsg
}

func getImagePullBackoffSolution(solution string) func(ctn *v1.Container, msg *string) []string {
	return func(ctn *v1.Container, msg *string) []string {
		return GetSolutionsByTemplate(solution, ctn, true)
	}
}
