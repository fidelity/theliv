/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package kubernetes

import (
	"strings"
	"testing"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

func TestCheckImagePullBackOffReason(t *testing.T) {
	for _, msg := range ImagePullBackOffReasons {
		assert.EqualValues(t, true, checkImagePullBackOffReason(msg))
		assert.EqualValues(t, true, checkImagePullBackOffReason(strings.ToUpper(msg)))
		assert.EqualValues(t, true, checkImagePullBackOffReason(strings.ToLower(msg)))
	}
	assert.EqualValues(t, false, checkImagePullBackOffReason(""))
	assert.EqualValues(t, false, checkImagePullBackOffReason("not-exist"))
}

var pod = v1.Pod{}
var Image = "a.b.com/c-123/busybox:v0.0.1"
var detectorInput = problem.DetectorCreationInput{
	Namespace:      "namespace",
	ClusterName:    "clustername",
	EventRetriever: nil,
	LogRetriever:   nil,
	Kubeconfig:     nil,
}
var problemInput = ProblemInput{
	Detector: &ResourceCommonDetector{
		DetectorInput: &detectorInput,
	},
}

func setup() {
	waiting := v1.ContainerStateWaiting{
		Reason:  ImagePullBackOffTitle,
		Message: "ImagePullBackOff message",
	}
	status := v1.PodStatus{
		ContainerStatuses: []v1.ContainerStatus{{
			State: v1.ContainerState{
				Waiting: &waiting,
			},
			Image: "imageName",
			Name:  "containerName",
		}},
	}
	pod.Status = status
	pod.Name = "A-pod"
}

func setupFullMessage() {
	pod.Status.ContainerStatuses[0].Image = Image
}

func TestGetSolutionUnknowManifestFew(t *testing.T) {
	setup()
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Either the image repository name is incorrect or does NOT exist.",
		"3. Either the image name is invalid or does NOT exist.",
		"4. Either the image tag is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnknowManifestNormal(t *testing.T) {
	setup()
	setupFullMessage()
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Either the image repository name c-123 is incorrect or does NOT exist.",
		"3. Either the image name busybox is invalid or does NOT exist.",
		"4. Either the image tag v0.0.1 is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnknowManifestDeepRepo(t *testing.T) {
	setup()
	pod.Status.ContainerStatuses[0].Image = "a.b.com/c-123/d-234/busybox:v0.0.1"
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'a.b.com/c-123/d-234/busybox:v0.0.1' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Either the image repository name c-123 is incorrect or does NOT exist.",
		"3. Either the image name busybox is invalid or does NOT exist.",
		"4. Either the image tag v0.0.1 is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnknowManifestShallowRepo(t *testing.T) {
	setup()
	pod.Status.ContainerStatuses[0].Image = "a.b.com/busybox:v0.0.1"
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'a.b.com/busybox:v0.0.1' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Either the image repository name is incorrect or does NOT exist.",
		"3. Either the image name is invalid or does NOT exist.",
		"4. Either the image tag is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnknowManifestImageMissing(t *testing.T) {
	setup()
	pod.Status.ContainerStatuses[0].Image = ""
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Either the image repository name is incorrect or does NOT exist.",
		"3. Either the image name is invalid or does NOT exist.",
		"4. Either the image tag is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnknowManifestContainerMissing(t *testing.T) {
	setup()
	pod.Status.ContainerStatuses[0].Name = ""
	res, _ := getSolutionUnknowManifest(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 5, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container." +
			" The root cause could be one of the following.",
		"2. Either the image repository name is incorrect or does NOT exist.",
		"3. Either the image name is invalid or does NOT exist.",
		"4. Either the image tag is invalid or does NOT exist.",
		"5. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionNoSuchHostFew(t *testing.T) {
	setup()
	res, _ := getSolutionNoSuchHost(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 2, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host is either incorrect or DNS is not able to resolve the hostname.",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionNoSuchHostFull(t *testing.T) {
	setup()
	setupFullMessage()
	res, _ := getSolutionNoSuchHost(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 2, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host a.b.com is either incorrect or DNS is not able to resolve the hostname.",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionIOTimeoutLess(t *testing.T) {
	setup()
	res, _ := getSolutionIOTimeout(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 3, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host is not reachable from kubelet in node because of a possible networking issue.",
		"3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionIOTimeoutFull(t *testing.T) {
	setup()
	setupFullMessage()
	pod.Spec.NodeName = "node-1"
	res, _ := getSolutionIOTimeout(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 3, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host a.b.com is not reachable from kubelet in node node-1 because of a possible networking issue.",
		"3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionConnectionRefusedLess(t *testing.T) {
	setup()
	res, _ := getSolutionConnectionRefused(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 3, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host is not reachable from kubelet in node because of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues.",
		"3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionConnectionRefusedFull(t *testing.T) {
	setup()
	setupFullMessage()
	pod.Spec.NodeName = "node-1"
	res, _ := getSolutionConnectionRefused(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 3, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry host a.b.com is not reachable from kubelet in node node-1 because of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues.",
		"3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnauthronizedLess(t *testing.T) {
	setup()
	res, _ := getSolutionUnauthronized(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 4, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. ImagePullSecret is either incorrect or expired.",
		"3. Repository does not exist.",
		"4. " + SecretMsg1NotExist,
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionUnauthronizedFull(t *testing.T) {
	setup()
	setupFullMessage()
	pod.Spec.ImagePullSecrets = []v1.LocalObjectReference{
		{
			Name: "test-secret",
		},
	}
	namespace := problemInput.Detector.DetectorInput.Namespace
	res, _ := getSolutionUnauthronized(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 7, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. ImagePullSecret test-secret is either incorrect or expired.",
		"3. Repository c-123 does not exist.",
		"4. " + SecretMsg1,
		"   kubectl get secret test-secret -n " + namespace +
			" --output=\"jsonpath={.data.\\.dockerconfigjson}\" | base64 --decode",
		"5. " + SecretMsg3,
		"   kubectl get pod " + pod.Name + " -n " + namespace + " --output=\"jsonpath={.spec." +
			"imagePullSecrets}\"",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionQuotaRateLimitLess(t *testing.T) {
	setup()
	res, _ := getSolutionQuotaRateLimit(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 2, len(res))
	expectedRes := []string{
		"1. Unable to pull image 'imageName' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry has been ratelimitted. Please increase the quota or limit.",
	}
	assert.EqualValues(t, expectedRes, res)
}

func TestGetSolutionQuotaRateLimitFull(t *testing.T) {
	setup()
	setupFullMessage()
	pod.Spec.NodeName = "node-1"
	res, _ := getSolutionQuotaRateLimit(&pod, &pod.Status.ContainerStatuses[0], &problemInput, nil, nil)
	assert.EqualValues(t, 2, len(res))
	expectedRes := []string{
		"1. Unable to pull image '" + Image + "' for the container 'containerName'." +
			" The root cause could be one of the following.",
		"2. Image registry a.b.com has been ratelimitted. Please increase the quota or limit.",
	}
	assert.EqualValues(t, expectedRes, res)
}
