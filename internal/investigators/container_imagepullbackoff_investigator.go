/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"strings"
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"
)

const (
	UnknownManifestMsg    = "The named manifest is not known"
	NoSuchHostMsg         = "No such host"
	IOTimeoutMsg          = "I/O timeout"
	NotFoundMsg           = "failed to pull and unpack image .* not found"
	ConnectionRefused     = "Connection refused"
	UnauthorizedMsg       = "Unauthorized or access denied or authentication required"
	AuthorizeFailed       = "failed to authorize"
	QuotaRateLimitMsg     = "Quota exceeded or Too Many Requests or rate limit"
	RepositoryNotExistMsg = "Repository does not exist or may require 'docker login'"
)

const (
	UnknownManifestSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. Either the image repository name is incorrect or does NOT exist.
3. Either the image name is invalid or does NOT exist.
4. Either the image tag is invalid or does NOT exist.
`
	NoSuchHostSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. Image registry host is either incorrect or DNS is not able to resolve the hostname.
`
	IOTimeoutSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. Image registry host is not reachable from kubelet in node {{ .Pod.Spec.NodeName}}, because of a possible networking issue.
3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)
`
	ConnectionRefusedSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. Image registry host is not reachable from kubelet in node {{ .Pod.Spec.NodeName}}, because of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues.
`
	UnauthorizedSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. ImagePullSecret is either incorrect or expired.
3. Repository does not exist.
`
	AuthorizeFailedSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. ImagePullSecret is either incorrect or expired.
3. Change ImagePullSecret to valid credential can fix this problem.
`
	QuotaRateLimitSolution = `
1. Unable to pull image {{ .Status.Image }} for the container {{ .Status.Name}}. The root cause could be one of the following.
2. Image registry has been rate limited. Please increase the quota or limit.
`
)

const (
	GetPoSecretCmd = `
1. kubectl get pod {{ .ObjectMeta.Name }} -n {{ .ObjectMeta.Namespace }} --output="jsonpath={.spec.imagePullSecrets}"
2. kubectl get secret {{ range .Spec.ImagePullSecrets }}{{ .Name }}{{ end }} -n {{ .ObjectMeta.Namespace }} --output="jsonpath={.data.\.dockerconfigjson}" | base64 --decode
`
	SecretMsg1 = `Run the following command 1 to make sure your imagepull secret is correct. Make sure under 'auths', a entry corresponding to you registry hostname exists with the CORRECT username & password. Sometime incorrect imagepull secret might lead to this error as well.
`
	SecretMsg2NotExist = `We see that pod does not reference to any imagePullSecret. Kindly make sure this was intentional and not missed by mistake. Imagepullsecrets are mandatory in order to pull image from a registry that requires authentication and does not support anonymous pulls
`
	SecretMsg3 = `Run the following command 2 to get the imagePullSecret name:
`
)

type PodAndStatus struct {
	Pod    v1.Pod
	Status *v1.ContainerStatus
}

var ImagePullBackOffSolutions = map[string]func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string){
	UnknownManifestMsg:    getImagePullBackOffSolution(UnknownManifestSolution),
	RepositoryNotExistMsg: getImagePullBackOffSolution(UnknownManifestSolution),
	NoSuchHostMsg:         getImagePullBackOffSolution(NoSuchHostSolution),
	IOTimeoutMsg:          getImagePullBackOffSolution(IOTimeoutSolution),
	ConnectionRefused:     getImagePullBackOffSolution(ConnectionRefusedSolution),
	UnauthorizedMsg:       getImagePullBackOffSolution(UnauthorizedSolution),
	AuthorizeFailed:       getImagePullBackOffSolution(AuthorizeFailedSolution),
	QuotaRateLimitMsg:     getImagePullBackOffSolution(QuotaRateLimitSolution),
	NotFoundMsg:           getImagePullBackOffSolution(UnknownManifestSolution),
}

var ImagePullBackOffReasons = []string{"ImagePullBackOff", "ErrImagePull", "ErrImagePullBackOff"}

func ContainerImagePullBackoffInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()

	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	for _, status := range pod.Status.ContainerStatuses {
		investigateContainerImgPullBackOff(ctx, problem, input, pod, status)
	}
}

func investigateContainerImgPullBackOff(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput, pod v1.Pod, status v1.ContainerStatus) {

	if status.State.Waiting != nil && checkImagePullBackOffReason(status.State.Waiting.Reason) {

		foundMsg := getPodSolutionFromEvents(ctx, problem, input, &pod, &status, ImagePullBackOffSolutions)

		if foundMsg == "" {
			detail := status.State.Waiting.Message
			solutions, _ := ImagePullBackOffSolutions[UnknownManifestMsg](ctx, &pod, &status)
			foundMsg = UnknownManifestMsg
			appendSolution(problem, detail, nil)
			appendSolution(problem, solutions, nil)
		}

		secretmsg := checksecretmsg(ctx, foundMsg, pod)
		appendSolution(problem, secretmsg, GetSolutionsByTemplate(ctx, GetPoSecretCmd, &pod, true))
	}
}

func checksecretmsg(ctx context.Context, msg string, pod v1.Pod) []string {
	var secretmsg []string
	if msg == UnknownManifestMsg || msg == RepositoryNotExistMsg || msg == NotFoundMsg {
		if len(pod.Spec.ImagePullSecrets) == 0 {
			s := "5. " + SecretMsg2NotExist
			secretmsg = GetSolutionsByTemplate(ctx, s, &pod, true)
		} else {
			s := "5. " + SecretMsg1 + "6. " + SecretMsg3
			secretmsg = GetSolutionsByTemplate(ctx, s, &pod, true)
		}
	} else if msg == UnauthorizedMsg || msg == AuthorizeFailed {
		if len(pod.Spec.ImagePullSecrets) == 0 {
			s := "4. " + SecretMsg2NotExist
			secretmsg = GetSolutionsByTemplate(ctx, s, &pod, true)
		} else {
			s := "4. " + SecretMsg1 + "5. " + SecretMsg3
			secretmsg = GetSolutionsByTemplate(ctx, s, &pod, true)
		}
	}
	return secretmsg
}

func checkImagePullBackOffReason(reason string) bool {
	// strings.Contains will return true when reason is empty string
	if len(reason) == 0 {
		return false
	}
	for _, msg := range ImagePullBackOffReasons {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(reason)) {
			return true
		}
	}
	return false
}

func getImagePullBackOffSolution(solution string) func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string) {
	return func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string) {

		s1 := GetSolutionsByTemplate(ctx, solution,
			PodAndStatus{
				Pod:    *pod,
				Status: status,
			}, true)
		return s1, nil
	}
}
