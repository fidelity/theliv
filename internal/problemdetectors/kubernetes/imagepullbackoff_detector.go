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

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*ImagePullBackoffDetector)(nil)

const (
	ImagePullBackOffDetectorName = "ImagePullBackOffDetector"
	UnknownManifestMsg           = "The named manifest is not known"
	NoSuchHostMsg                = "No such host"
	IOTimeoutMsg                 = "I/O timeout"
	ConnectionRefused            = "Connection refused"
	UnauthronizedMsg             = "Unauthorized or access denied or authentication required"
	QuotaRateLimitMsg            = "Quota exceeded or Too Many Requests or rate limit"
	RepositoryNotExistMsg        = "Repository does not exist or may require 'docker login'"
	ImagePullBackOffDocLink      = "https://kubernetes.io/docs/concepts/containers/images/#imagepullbackoff"
	ImagePullBackOffTitle        = "ImagePullBackOff"
)

const (
	SecretMsg1 = "Run the following command to make sure your imagepull secret is correct. Make sure under " +
		"'auths', a entry corresponding to you registry hostname exists with the CORRECT username & password. " +
		"Sometime incorrect imagepull secret might lead to this error as well."
	SecretMsg1NotExist = "We see that pod does not reference to any imagePullSecret. Kindly make sure this was " +
		"intentional and not missed by mistake. Imagepullsecrets are mandatory in order to " +
		"pull image from a registry that requires authentication and does not support anonymous pulls"
	SecretMsg2 = "   kubectl get secret %s-n %s --output=\"jsonpath={.data.\\.dockerconfigjson}\" | base64 --decode"
	SecretMsg3 = "Run the following command to get the imagePullSecret name"
	SecretMsg4 = "   kubectl get pod %s -n %s --output=\"jsonpath={.spec.imagePullSecrets}\""
)

var ImagePullBackoffTags = []string{"imagepullbackoff", "kubelet"}
var ImagePullBackOffReasons = []string{"ImagePullBackOff", "ErrImagePull", "ErrImagePullBackOff"}

var ImagePullBackoffSolutions = map[string]func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	msg *string, e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL){
	UnknownManifestMsg:    getSolutionUnknowManifest,
	RepositoryNotExistMsg: getSolutionUnknowManifest,
	NoSuchHostMsg:         getSolutionNoSuchHost,
	IOTimeoutMsg:          getSolutionIOTimeout,
	ConnectionRefused:     getSolutionConnectionRefused,
	UnauthronizedMsg:      getSolutionUnauthronized,
	QuotaRateLimitMsg:     getSolutionQuotaRateLimit,
}

// Depends on registry, the error message may be different, use the possible messages to check.
var PossibleErrorMessages = map[string][]string{
	UnknownManifestMsg: {"manifest unknown", "manifest not available"},
}

func RegisterImagePullBackOffWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {
	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(ImagePullBackOffDetectorName),
			Description:   "This detector will detect ImagePullBackoff error",
			Documentation: `Container image pull failed, kubelet is backing off image pull`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewImagePullBackoff,
	}, problem.PodFailuresDomain)
	return err
}

func NewImagePullBackoff(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return ImagePullBackoffDetector{
		ResourceCommonDetector{
			name:          ImagePullBackOffDetectorName,
			DetectorInput: i,
		}}, nil
}

type ImagePullBackoffDetector struct {
	ResourceCommonDetector
}

// Check all pods not in running/succeeded state, and pod failure reason is ImagePullBackOff.
// When event client is provided, check the events to figure out root cause, without events, the detector
// can not determine the root cause.
// The problem with same root cause will be consolidated into one problem.Problem, the corresponding pods will be
// added to AffectedResources.
func (d ImagePullBackoffDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	problemInput := &ProblemInput{
		Detector:              &d.ResourceCommonDetector,
		PodSkipDetect:         podSkipDetectImagePullBackOff,
		PodDetect:             podDetectImagePullBackOff,
		SolutionsLinks:        ImagePullBackoffSolutions,
		Title:                 ImagePullBackOffTitle,
		Tags:                  ImagePullBackoffTags,
		DocLink:               ImagePullBackOffDocLink,
		PossibleErrorMessages: PossibleErrorMessages,
	}
	return PodsDetect(ctx, problemInput)
}

func getImageContainerName(msg string, image string, container string) string {
	imageStr := image
	if len(image) > 0 {
		imageStr = " '" + imageStr + "'"
	}
	containerStr := container
	if len(container) > 0 {
		containerStr = " '" + containerStr + "'"
	}
	return fmt.Sprintf(msg, imageStr, containerStr)
}

func getSolutionUnknowManifest(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. Either the image repository name %sis incorrect or does NOT exist."
	msg3 := "3. Either the image name %sis invalid or does NOT exist."
	msg4 := "4. Either the image tag %sis invalid or does NOT exist."

	imageDetails := strings.Split(status.Image, "/")
	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))
	if len(imageDetails) > 2 {
		msg2 = fmt.Sprintf(msg2, imageDetails[1]+" ")
		imageVersion := strings.Split(imageDetails[len(imageDetails)-1], ":")
		if len(imageVersion) > 1 {
			msg3 = fmt.Sprintf(msg3, imageVersion[0]+" ")
			msg4 = fmt.Sprintf(msg4, imageVersion[1]+" ")
		} else {
			msg3 = fmt.Sprintf(msg3, "")
			msg4 = fmt.Sprintf(msg4, "")
		}
	} else {
		msg2 = fmt.Sprintf(msg2, "")
		msg3 = fmt.Sprintf(msg3, "")
		msg4 = fmt.Sprintf(msg4, "")
	}
	secretsLength := len(pod.Spec.ImagePullSecrets)
	msg5, msg6, msg8 := "5. "+SecretMsg1, "", ""
	if secretsLength == 0 {
		msg5 = "5. " + SecretMsg1NotExist
	} else if secretsLength > 0 {
		// In most cases, there is only one secret, pick the first in the displayed command if
		// there is more than one secrets.
		msg6 = fmt.Sprintf(SecretMsg2, pod.Spec.ImagePullSecrets[0].Name+" ",
			problemInput.Detector.DetectorInput.Namespace)
	}
	msg8 = fmt.Sprintf(SecretMsg4, pod.Name, problemInput.Detector.DetectorInput.Namespace)

	var solutions []string
	if secretsLength == 0 {
		solutions = []string{msg1, msg2, msg3, msg4, msg5}
	} else {
		solutions = []string{msg1, msg2, msg3, msg4, msg5, msg6, "6. " + SecretMsg3, msg8}
	}

	return solutions, getLogEventLinks(pod, problemInput, true, true, false, e)
}

func getSolutionNoSuchHost(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. Image registry host %sis either incorrect or DNS is not able to resolve the hostname."

	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))
	imageDetails := strings.Split(status.Image, "/")
	// If there is one item, it means the image does not contain slash, skip this case.
	if len(imageDetails) > 1 {
		msg2 = fmt.Sprintf(msg2, imageDetails[0]+" ")
	} else {
		msg2 = fmt.Sprintf(msg2, "")
	}
	return []string{msg1, msg2}, getLogEventLinks(pod, problemInput, true, true, false, e)
}

func getSolutionIOTimeout(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. Image registry host %sis not reachable from kubelet in node %sbecause of a possible networking issue."
	msg3 := "3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)"

	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))

	imageDetails := strings.Split(status.Image, "/")
	if len(imageDetails) > 1 {
		msg2 = fmt.Sprintf(msg2, imageDetails[0]+" ", pod.Spec.NodeName+" ")
	} else {
		msg2 = fmt.Sprintf(msg2, "", "")
	}
	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, false, e)
}

func getSolutionConnectionRefused(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. Image registry host %sis not reachable from kubelet in node %sbecause of a possible networking issue. Please check your firewall rules to make sure connection is not being refused by any firewall. Sometimes this could be due to intermittent n/w issues."
	msg3 := "3. Make sure your subnet has reachability to Amazon ECR. Make sure your worker node security group has access allowed to reach Amazon ECR (if aws, similarly for other cloud providers)"

	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))
	imageDetails := strings.Split(status.Image, "/")
	if len(imageDetails) > 1 {
		msg2 = fmt.Sprintf(msg2, imageDetails[0]+" ", pod.Spec.NodeName+" ")
	} else {
		msg2 = fmt.Sprintf(msg2, "", "")
	}
	return []string{msg1, msg2, msg3}, getLogEventLinks(pod, problemInput, true, true, false, e)
}

func getSolutionUnauthronized(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. ImagePullSecret %sis either incorrect or expired."
	msg3 := "3. Repository %sdoes not exist."

	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))

	secrects := pod.Spec.ImagePullSecrets
	if len(secrects) > 0 {
		secrectsName := ""
		for _, secret := range secrects {
			secrectsName += (secret.Name + ",")
		}
		if len(secrectsName) > 0 {
			msg2 = fmt.Sprintf(msg2, secrectsName[:len(secrectsName)-1]+" ")
		} else {
			msg2 = fmt.Sprintf(msg2, "")
		}
	} else {
		msg2 = fmt.Sprintf(msg2, "")
	}
	imageDetails := strings.Split(status.Image, "/")
	if len(imageDetails) > 1 {
		msg3 = fmt.Sprintf(msg3, imageDetails[1]+" ")
	} else {
		msg3 = fmt.Sprintf(msg3, "")
	}

	secretsLength := len(pod.Spec.ImagePullSecrets)
	msg4, msg5, msg7 := "4. "+SecretMsg1, "", ""
	if secretsLength == 0 {
		msg4 = "4. " + SecretMsg1NotExist
	} else if secretsLength > 0 {
		msg5 = fmt.Sprintf(SecretMsg2, pod.Spec.ImagePullSecrets[0].Name+" ",
			problemInput.Detector.DetectorInput.Namespace)
	}
	msg7 = fmt.Sprintf(SecretMsg4, pod.Name, problemInput.Detector.DetectorInput.Namespace)

	var solutions []string
	if secretsLength == 0 {
		solutions = []string{msg1, msg2, msg3, msg4}
	} else {
		solutions = []string{msg1, msg2, msg3, msg4, msg5, "5. " + SecretMsg3, msg7}
	}

	return solutions, getLogEventLinks(pod, problemInput, true, true, false, e)
}

func getSolutionQuotaRateLimit(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	msg1 := "1. Unable to pull image%s for the container%s. The root cause could be one of the following."
	msg2 := "2. Image registry %shas been ratelimitted. Please increase the quota or limit."

	msg1 = getImageContainerName(msg1, status.Image, getContainerName(pod, ImagePullBackOffTitle))
	imageDetails := strings.Split(status.Image, "/")
	if len(imageDetails) > 1 {
		msg2 = fmt.Sprintf(msg2, imageDetails[0]+" ")
	} else {
		msg2 = fmt.Sprintf(msg2, "")
	}
	return []string{msg1, msg2}, getLogEventLinks(pod, problemInput, true, true, false, e)
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

func podSkipDetectImagePullBackOff(po v1.Pod) bool {
	return po.Status.Phase == v1.PodRunning || po.Status.Phase == v1.PodSucceeded
}

func podDetectImagePullBackOff(status v1.ContainerStatus) bool {
	return status.State.Waiting != nil && checkImagePullBackOffReason(status.State.Waiting.Reason)
}
