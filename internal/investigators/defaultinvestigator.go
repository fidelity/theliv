package investigators

import (
	"context"
	golog "log"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DetaultInvestigator(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	switch problem.Tags["resourcetype"] {
	case "pod":
		getPodDetails(ctx, problem, input)
	case "container":
		getContainerDetails(ctx, problem, input)
	case "deployment":
		getDeploymentDetails(ctx, problem, input)
	}
}

func getPodDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	pod := &v1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)

	golog.Printf("INFO - Checking status with pod %s", pod.Name)
	if pod.Status.Message == "" && pod.Status.Reason == "" {
		return
	}
	detail := buildReasonMsg(pod.Status.Reason, pod.Status.Message)
	problem.Details = append(problem.Details, &detail)
}

func getContainerDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	containername := problem.Tags["container"]
	pod := &v1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)
	golog.Printf("INFO - Checking init container with pod %s", pod.Name)
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.Details = append(problem.Details, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.Details = append(problem.Details, &detail)
			}
			break
		}
	}

	golog.Printf("INFO - Checking container with pod %s", pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.Details = append(problem.Details, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.Details = append(problem.Details, &detail)
			}
			break
		}
	}

}

func getDeploymentDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["deployment"],
	}
	deployment := &appsv1.Deployment{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, deployment, namespace, getOptions)

	golog.Printf("INFO - Checking status with deployment %s", deployment.Name)
	for _, condition := range deployment.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message == "" && condition.Reason == "" {
			break
		}
		detail = detail + buildReasonMsg(condition.Reason, condition.Message)
		problem.Details = append(problem.Details, &detail)
	}

}

func buildReasonMsg(reason string, message string) string {
	detail := ""
	if reason != "" {
		detail = detail + "Reason: " + reason + ". "
	}
	if message != "" {
		detail = detail + "Message: " + message + ". "
	}
	return detail
}
