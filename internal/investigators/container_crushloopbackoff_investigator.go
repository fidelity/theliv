/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"fmt"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"
)

const (
	CrashLoopBackOff = "CrashLoopBackOff"

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

	CrushLoopBackOffMsg     = "Container {{.ContainerName}} has been restarted more than 10 times in the last few minutes."
	KubectlLogCmd           = "cmd: kubectl logs {{.Pod.Name}} -p -c {{.ContainerName}} -n {{.Pod.ObjectMeta.Namespace}}"
	KubectlDescribePoCmd    = "cmd: kubectl describe po {{.Pod.Name}} -n {{.Pod.ObjectMeta.Namespace}}"
	CrashLoopBackOffDocLink = "https://containersolutions.github.io/runbooks/posts/kubernetes/crashloopbackoff"
)

const (
	SolutionExitCode1 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code (1). General exit with errors.
Check your command or application logs.
`
	SolutionExitCode126 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code (126). Command invoked cannot execute.
Check your command, may be Permission problem or command is not an executable.
`
	SolutionExitCode127 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code (127). Command not found.
Check your command, maybe executable not in $PATH, or file not found.
`
	SolutionExitCode2To128 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code {{.ExitCode}}. May caused by application in container.
Check your command or application logs.
`
	SolutionExitCode137 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code (137). Container was killed.
Generally caused by OOM.
`
	SolutionExitCode139 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code (139). Errors in container.
Issues can be in your application codes or the base image, Check your dockerFile or application logs.
`
	SolutionExitCode129To255 = `
Container {{.ContainerName}} has EXITED with with a non-zero exit code {{.ExitCode}}. Container was terminated from external.
Check your application logs or system configurations.
`
	SolutionOOM = `
Container {{.ContainerName}} has EXITED with reason OOMKilled (1). Check the resource limits of the container.
`
	SolutionReadinessProbeFail = `
Following readiness probe has failed for the container {{.ContainerName}}.
`
	SolutionLivenessProbeFail = `
Following liveliness probe has failed for the container {{.ContainerName}}.
`
	SolutionStartupProbeFailMsg = `
Following startup probe has failed for the container {{.ContainerName}}.
`
	SolutionExecutableNotFoundMsg = `
Container {{.ContainerName}} has EXITED with a non-zero exit code (127). Check your command or application startup logs.
Give more insights in the UI based on this https://intl.cloud.tencent.com/document/product/457/35758. E.g if exit code is 127, then look at your \"command\" and make sure it is correct. Problem is there.
Container {{.ContainerName}} was unable to start, logs can be retrieved by 1 of the following 2 steps.
`
	SolutionNoSuchFile = `
Container {{.ContainerName}} has EXITED with a non-zero exit code (127). Check your commands or application startup logs.
Give more insights in the UI based on this https://intl.cloud.tencent.com/document/product/457/35758. E.g if exit code is 127, then look at your \"command\" and make sure it is correct. Problem is there.
Container {{.ContainerName}} was unable to start, logs can be retrieved by 1 of the following 2 steps.
`
	DefaultSolution = CrushLoopBackOffMsg + "\n" + CrashLoopBackOffDocLink + "\n" + KubectlDescribePoCmd + "\n"
)

type CrushLoopPodInfo struct {
	Pod           v1.Pod
	ContainerName string
	ExitCode      int32
}

var CrashLoopBackOffSolutions = map[string]func(pod *v1.Pod, status *v1.ContainerStatus) []string{
	ExecutableNotFoundMsg: getCrushLoopBackOffCommonSolution(SolutionExecutableNotFoundMsg, nil),
	NoSuchFileMsg:         getCrushLoopBackOffCommonSolution(SolutionNoSuchFile, nil),
	ReadinessProbeFailMsg: getCrushLoopBackOffCommonSolution(CrushLoopBackOffMsg+SolutionReadinessProbeFail, getSolutionReadinessProbeFailMsg),
	LivenessProbeFailMsg:  getCrushLoopBackOffCommonSolution(CrushLoopBackOffMsg+SolutionLivenessProbeFail, getSolutionLivenessProbeFailMsg),
	StartupProbeFailMsg:   getCrushLoopBackOffCommonSolution(CrushLoopBackOffMsg+SolutionStartupProbeFailMsg, getSolutionStartupProbeFailMsg),
	ExitWithOOM:           getCrushLoopBackOffCommonSolution(CrushLoopBackOffMsg+SolutionOOM, getSolutionOOM),
	ExitCode1:             getCrushLoopBackOffCommonSolution(SolutionExitCode1, nil),
	ExitCode126:           getCrushLoopBackOffCommonSolution(SolutionExitCode126, nil),
	ExitCode127:           getCrushLoopBackOffCommonSolution(SolutionExitCode127, nil),
	ExitCode2To128:        getCrushLoopBackOffCommonSolution(SolutionExitCode2To128, nil),
	ExitCode137:           getCrushLoopBackOffCommonSolution(SolutionExitCode137, nil),
	ExitCode139:           getCrushLoopBackOffCommonSolution(SolutionExitCode139, nil),
	ExitCode129To255:      getCrushLoopBackOffCommonSolution(SolutionExitCode129To255, nil),
}

func ContainerCrashLoopBackoffInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {

	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason == CrashLoopBackOff {
			if getPodSolutionFromEvents(ctx, problem, input, &pod, &status, CrashLoopBackOffSolutions) == "" {
				code := getRootCauseByExitCode(&pod)
				if code != "" {
					addSolutionFromMap(problem, &pod, nil, code, CrashLoopBackOffSolutions)
					return
				}
				solution := getCrushLoopBackOffCommonSolution(DefaultSolution, nil)(&pod, &v1.ContainerStatus{})
				appendSolution(problem, solution)
			}
		}
	}
}

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

func getCrushLoopBackOffCommonSolution(template string, addStep func(pod *v1.Pod) string) func(pod *v1.Pod, status *v1.ContainerStatus) []string {
	return func(pod *v1.Pod, status *v1.ContainerStatus) []string {

		containerName := getContainerName(pod, CrashLoopBackOff)
		containerStatus := getFailedContainerStatus(pod, containerName)
		code := containerStatus.LastTerminationState.Terminated.ExitCode

		solution := GetSolutionsByTemplate(CrushLoopBackOffMsg+template+KubectlLogCmd,
			getPodInfo(*pod, containerName, code), true)

		if addStep != nil {
			solution = append(solution, addStep(pod))
		}

		return solution
	}
}

func getSolutionOOM(pod *v1.Pod) string {

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	return getResourceLimit(&container.Resources)
}

func getSolutionReadinessProbeFailMsg(pod *v1.Pod) string {

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	return getProbe(container.ReadinessProbe, "readiness")
}

func getSolutionLivenessProbeFailMsg(pod *v1.Pod) string {

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	return getProbe(container.LivenessProbe, "liveness")
}

func getSolutionStartupProbeFailMsg(pod *v1.Pod) string {

	containerName := getContainerName(pod, CrashLoopBackOff)
	container := getFailedContainer(pod, containerName)

	return getProbe(container.StartupProbe, "startup")
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

// Return Container's Probe definition, string of multiple lines.
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

// Return Container's resource requests/limits definition, string of multiple lines.
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

func getContainerName(pod *v1.Pod, reason string) string {
	if pod == nil {
		return ""
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason == reason {
			return status.Name
		}
	}
	return ""
}

func getPodInfo(pod v1.Pod, containerName string, code int32) CrushLoopPodInfo {
	return CrushLoopPodInfo{
		Pod:           pod,
		ContainerName: containerName,
		ExitCode:      code,
	}
}
