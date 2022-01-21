/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package kubernetes

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/fidelity/theliv/internal/problem"
	observability "github.com/fidelity/theliv/pkg/observability"

	v1 "k8s.io/api/core/v1"
)

var _ problem.Detector = (*CrashLoopBackOffDetector)(nil)

const (
	CrashLoopBackOff             = "CrashLoopBackOff"
	CrashLoopBackOffDetectorName = CrashLoopBackOff + "Detector"

	ExecutableNotFoundMsg = "executable file not found in $PATH: unknown"
	NoSuchFileMsg         = "no such file or directory: unknown"
	ReadinessProbeFailMsg = "Readiness probe failed"
	LivenessProbeFailMsg  = "Liveness probe failed"
	StartupProbeFailMsg   = "Startup probe failed"
	ExitWithOOM           = "Container killed due to OOM."
	ExitCode1             = "Generally application has issues, container exit code(1)."
	ExitCode126           = "Command invoked cannot execute, container exit code(126)."
	ExitCode127           = "Command not found, container exit code(127)."
	ExitCode2To128        = "Container terminated internally, EXIT with Non-Zero value(between 2 to 128)."
	ExitCode137           = "Container was killed, generally due to OOM, container exit code(137)."
	ExitCode139           = "Issues in application code or the base image, container exit code(139)"
	ExitCode129To255      = "Container terminated by external signal, EXIT with Non-Zero value(between 129 to 255)."
	Step1Msg              = "1. Container '%s' has been restarted more than 10 times in the last few minutes."

	CrashLoopBackOffDocLink = "https://containersolutions.github.io/runbooks/posts/kubernetes/crashloopbackoff"
	KubectlLogCmd           = "cmd: kubectl logs %s -p -c %s -n %s"
)

var CrashLoopBackOffTags = []string{strings.ToLower(CrashLoopBackOff), "kubelet"}

var CrashLoopBackOffSolutions = map[string]func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	msg *string, e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL){
	ExecutableNotFoundMsg: getSolutionExecutableNotFoundMsg,
	NoSuchFileMsg:         getSolutionNoSuchFile,
	ReadinessProbeFailMsg: getSolutionReadinessProbeFailMsg,
	LivenessProbeFailMsg:  getSolutionLivenessProbeFailMsg,
	StartupProbeFailMsg:   getSolutionStartupProbeFailMsg,
	ExitWithOOM:           getSolutionOOM,
	ExitCode1:             getSolutionExitCode1,
	ExitCode126:           getSolutionExitCode126,
	ExitCode127:           getSolutionExitCode127,
	ExitCode2To128:        getSolutionExitCode2To128,
	ExitCode137:           getSolutionExitCode137,
	ExitCode139:           getSolutionExitCode139,
	ExitCode129To255:      getSolutionExitCode129To255,
}

func RegisterCrashLoopBackOffWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {

	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(CrashLoopBackOffDetectorName),
			Description:   "This detector will detect CrashLoopBackOff error",
			Documentation: `Container crashed, kubelet is backing off container crash`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewCrashLoopBackOff,
	}, problem.PodFailuresDomain)
	return err
}

func NewCrashLoopBackOff(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return CrashLoopBackOffDetector{
		ResourceCommonDetector{
			name:          CrashLoopBackOffDetectorName,
			DetectorInput: i,
		}}, nil
}

type CrashLoopBackOffDetector struct {
	ResourceCommonDetector
}

// Check all pods not in running/successed state, and pod failure reason is CrashLoopBackOff.
// When event client is provided, check the events to figure out root cause
// If no root cause found from events, use the lastTerminated, exit code and exit reason to determine the possible root cause.
// The problem with same root cause will be consolidted into one problem.Problem, the corresponding pods will be
// added to AffectedResources.
func (d CrashLoopBackOffDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	problemInput := &ProblemInput{
		Detector:       &d.ResourceCommonDetector,
		PodSkipDetect:  podSkipDetectCrashLoopBackOff,
		PodDetect:      podDetectCrashLoopBackOff,
		SolutionsLinks: CrashLoopBackOffSolutions,
		Title:          CrashLoopBackOff,
		Tags:           CrashLoopBackOffTags,
		DocLink:        CrashLoopBackOffDocLink,
	}
	return PodsDetect(ctx, problemInput)
}

func podSkipDetectCrashLoopBackOff(po v1.Pod) bool {
	return (po.Status.Phase == v1.PodRunning && podReadyCheck(po.Status.Conditions)) || po.Status.Phase == v1.PodSucceeded
}

func podDetectCrashLoopBackOff(status v1.ContainerStatus) bool {
	return status.State.Waiting != nil && status.State.Waiting.Reason == CrashLoopBackOff
}

// Use container lastTerminated, container exit code and exit reason to determin the possible root cause.
// Exit Code 1, general non-zero exit.
// Exit Code 126, command invoked can't be executed.
// Exit Code 127, command not found.
// Exit Code 137, container was killed, generally due to OOM.
// Exit Code 1 to 128, container terminated internally.
// Exit Code 129 to 255, container terminated by external signal.
func getRootCauseByExitCode(pod *v1.Pod) string {
	containerName := getContainerName(pod, CrashLoopBackOff)
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName && container.LastTerminationState.Terminated != nil {
			exitCode := container.LastTerminationState.Terminated.ExitCode
			reason := container.LastTerminationState.Terminated.Reason
			switch {
			case exitCode == 1:
				if reason == "OOMKilled" {
					return ExitWithOOM
				} else {
					return ExitCode1
				}
			case exitCode == 126:
				return ExitCode126
			case exitCode == 127:
				return ExitCode127
			case exitCode > 1 && exitCode <= 128:
				return ExitCode2To128
			case exitCode == 137:
				return ExitCode137
			case exitCode >= 129 && exitCode < 255:
				return ExitCode129To255
			case exitCode == 255:
				return ExitCode1
			}
		}
	}
	return ""
}

func getSolutionOOM(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with reason OOMKilled (1). Check the resource limits of the container."

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg3 := getResourceLimit(&container.Resources)
	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, true, nil)
}

func getSolutionExitCode1(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (1). General exit with errors."
	msg3 := "3. Check your command or application logs."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)

	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode126(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (126). Command invoked cannot execute."
	msg3 := "3. Check your command, may be Permission problem or command is not an executable."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)
	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode127(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (127). Command not found."
	msg3 := "3. Check your command, maybe executable not in $PATH, or file not found."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)
	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode2To128(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (%d). May caused by application in container."
	msg3 := "3. Check your command or application logs."

	containerName := getContainerName(pod, CrashLoopBackOff)
	containerStatus := getFailedContainerStatus(pod, containerName)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName, containerStatus.LastTerminationState.Terminated.ExitCode)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)

	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode137(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (137). Container was killed."
	msg3 := "3. Generally caused by OOM."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)

	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode139(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (139). Errors in container."
	msg3 := "3. Issues can be in your application codes or the base image, Check your dockerFile or application logs."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)

	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExitCode129To255(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with with a non-zero exit code (%d). Container was terminated from external."
	msg3 := "3. Check your application logs or system configurations."

	containerName := getContainerName(pod, CrashLoopBackOff)
	containerStatus := getFailedContainerStatus(pod, containerName)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName, containerStatus.LastTerminationState.Terminated.ExitCode)
	msg4 := fmt.Sprintf("4. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)
	return []string{msg1, msg2, msg3, msg4}, getLogEventLinks(pod, problemInput, false, false, true, nil)
}

func getSolutionExecutableNotFoundMsg(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with a non-zero exit code (127). Check your command or application startup logs."
	msg3 := "3. Give more insights in the UI based on this https://intl.cloud.tencent.com/document/product/457/35758. E.g if exit code is 127, then look at your \"command\" and make sure it is correct. Problem is there."
	msg4 := "4. Container '%s' was unable to start, logs can be retrieved by 1 of the following 2 steps."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 = fmt.Sprintf(msg4, containerName)
	msg5 := fmt.Sprintf("5. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)
	return []string{msg1, msg2, msg3, msg4, msg5}, getLogEventLinks(pod, problemInput, false, false, true, e)
}

func getSolutionNoSuchFile(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Container '%s' has EXITED with a non-zero exit code (127). Check your commands or application startup logs."
	msg3 := "3. Give more insights in the UI based on this https://intl.cloud.tencent.com/document/product/457/35758. E.g if exit code is 127, then look at your \"command\" and make sure it is correct. Problem is there."
	msg4 := "4. Container '%s' was unable to start, logs can be retrieved by 1 of the following 2 steps."

	containerName := getContainerName(pod, CrashLoopBackOff)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg4 = fmt.Sprintf(msg4, containerName)
	msg5 := fmt.Sprintf("5. "+KubectlLogCmd, pod.Name, containerName, pod.Namespace)

	return []string{msg1, msg2, msg3, msg4, msg5}, getLogEventLinks(pod, problemInput, false, false, true, e)
}

func getSolutionReadinessProbeFailMsg(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Following readiness probe has failed for the container '%s'."

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)
	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg3 := getProbe(container.ReadinessProbe, "readiness")

	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, true, e)
}

func getSolutionLivenessProbeFailMsg(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Following liveliness probe has failed for the container '%s'."

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg3 := getProbe(container.LivenessProbe, "liveness")
	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, true, e)
}

func getSolutionStartupProbeFailMsg(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg2 := "2. Following startup probe has failed for the container '%s'."

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	msg1 := fmt.Sprintf(Step1Msg, containerName)
	msg2 = fmt.Sprintf(msg2, containerName)
	msg3 := getProbe(container.StartupProbe, "startup")
	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, true, e)
}

func podReadyCheck(conditon []v1.PodCondition) bool {
	for _, con := range conditon {
		if con.Type == "Ready" && con.Status == "True" {
			return true
		}
	}
	return false
}

func getFailedContainer(pod *v1.Pod, containerName string) v1.Container {
	for _, container := range pod.Spec.Containers {
		if container.Name == containerName {
			return container
		}
	}
	return pod.Spec.Containers[0]
}

func getFailedContainerStatus(pod *v1.Pod, containerName string) v1.ContainerStatus {
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName {
			return container
		}
	}
	return pod.Status.ContainerStatuses[0]
}

// Return Container's Probe defination, string of multiple lines.
func getProbe(probe *v1.Probe, probeType string) string {

	if probe == nil {
		return ""
	}

	msg := `%sProbe:
failureThreshold: %d
initialDelaySeconds: %d
periodSeconds: %d
successThreshold: %d
timeoutSeconds: %d`

	msg = fmt.Sprintf(msg, probeType, probe.FailureThreshold, probe.InitialDelaySeconds, probe.PeriodSeconds,
		probe.SuccessThreshold, probe.TimeoutSeconds)

	if probe.HTTPGet != nil {
		http := `
httpGet:
path: %s
port: %d
scheme: %s`
		http = fmt.Sprintf(http, probe.HTTPGet.Path, probe.HTTPGet.Port.IntVal, fmt.Sprint(probe.HTTPGet.Scheme))
		msg = msg + http
	}
	return msg
}

// Return Container's resource requests/limits defination, string of multiple lines.
func getResourceLimit(resource *v1.ResourceRequirements) string {

	if resource == nil {
		return ""
	}
	var msg string
	if resource.Requests != nil {
		msg1 := `
Requests:
cpu: %s
memory: %s`
		msg1 = fmt.Sprintf(msg1, resource.Requests.Memory(), resource.Requests.Memory())
		msg = msg + msg1
	}
	if resource.Limits != nil {
		msg1 := `
Limits:
cpu: %s
memory: %s`
		msg1 = fmt.Sprintf(msg1, resource.Limits.Memory(), resource.Limits.Memory())
		msg = msg + msg1
	}

	return msg
}
