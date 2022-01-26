package ingress

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/fidelity/theliv/pkg/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	tag "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/internal/problemdetectors/kubernetes"
	"github.com/fidelity/theliv/pkg/csp/awsclient"
	"github.com/fidelity/theliv/pkg/kubeclient"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*IngressEksDetector)(nil)
var wg sync.WaitGroup

const (
	IngressEksDetectorName  = "IngressEksDetector"
	IngressEksTitle         = "IngressEks"
	IngressEksDocLink       = "https://docs.aws.amazon.com/eks/latest/userguide/alb-ingress.html"
	InvalidAnnotation       = "Invalid annotations for ALB ingress controller"
	InvalidAnnotationTitle  = "Invalid annotations"
	InvalidAnnotationDesc   = "The annotations in the ingress are invalid annotations for ALB ingress controller"
	InvalidAnnotationDoc    = "https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/"
	InvalidIngressPathTitle = "Invalid ingress rule"
	InvalidIngressPathDesc  = "The service name or port is not matched with service resource"
	InvalidIngressPathDoc   = "https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-rules"
	SSLAnnotation           = "alb.ingress.kubernetes.io/certificate-arn"
	SSLAnnotationTitle      = "Invalid SSL annotation"
	SSLAnnotationDesc       = "The annotations in the ingress are invalid annotations for ALB ingress controller"
	SSLAnnotationDoc        = "https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#certificate-arn"
	SSLSeparator            = ","
	InvalidSSLCertTitle     = "Invalid ingress certifivate-arn"
	InvalidSSLCertDesc      = "Invalid hostname in ingress"
	InvalidSSLCertDoc       = "https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/#certificate-arn"
	ExpiredSSLCertTitle     = "Certificate expired"
	InactiveSSLCertTitle    = "Certificate inactive"
	ExpiredSSLCertDesc      = "Certificate specified in alb.ingress.kubernetes.io/certificate-arn is expired"
	ExpiredSSLCertDoc       = "https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-describe.html"
	ALBIssueTitle           = "ALB is not in active status"
	ALBIssueDesc            = "Application load balancer is not in active status"
	ALBIssueDoc             = "https://docs.aws.amazon.com/elasticloadbalancing/latest/application/introduction.html"
	ALBNotExistTitle        = "ALB is not created"
	ALBNotExistDesc         = "Application load balancer is not created"
	ALBAnnotationPrefix     = "alb.ingress.kubernetes.io"
	TargetGroupIssueTitle   = "ALB target is in unhealthy status"
	TargetGroupIssueDesc    = "The target in ALB's target group is in unhealthy status, please note if all targets are unhealthy, the traffic will be impacted."
	TargetGroupIssueDoc     = "https://aws.amazon.com/premiumsupport/knowledge-center/elb-fix-failing-health-checks-alb/"
	HostnameWildcard        = "*"
)

var InvalidIngressTags = []string{"ingress"}

func RegisterIngressEksWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {
	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(IngressEksDetectorName),
			Description:   "This detector will detect IngressEks error",
			Documentation: "",
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform},
		},
		CreateFunc: NewIngressEksDetector,
	}, problem.IngressFailuresDomain)
	return err
}

func NewIngressEksDetector(input *problem.DetectorCreationInput) (problem.Detector, error) {
	return IngressEksDetector{
		name:          IngressEksDetectorName,
		DetectorInput: input,
	}, nil
}

type IngressEksDetector struct {
	DetectorInput *problem.DetectorCreationInput
	name          string
}

type TargetGroupTags struct {
	arn       *string
	tagOutput *tag.GetResourcesOutput
}

// Detect ingress issues, including annotation, ingress/service mapping, AWS related resource,
// application load balancer, target group and SSL certificate valid period.
func (d IngressEksDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	client, err := kubeclient.NewKubeClient(d.DetectorInput.Kubeconfig)
	if err != nil {
		log.S().Errorf("Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: d.DetectorInput.Namespace,
	}

	ingressList := &network.IngressList{}
	ingressListOptions := metav1.ListOptions{}
	client.List(ctx, ingressList, namespace, ingressListOptions)
	problems := make([]problem.Problem, 0)
	ingresses := []network.Ingress{}

	if len(ingressList.Items) == 0 {
		log.S().Info("No ingress found in namespace")
	} else {
		checkAnnotation(d.Domain(), namespace.Namespace, ingressList, &ingresses, &problems)

		checkService(ctx, d.Domain(), namespace, client, &ingresses, &problems)

		checkAlb(ctx, d.Domain(), d.DetectorInput, namespace.Namespace, &ingresses, &problems)

		checkTargetGroups(ctx, d.Domain(), namespace.Namespace, d.DetectorInput.AwsConfig, &ingresses, &problems)

		// eksClusterInfo, err := GetEksClusterInfo(ctx, d.DetectorInput.AwsConfig, d.DetectorInput.ClusterName)
		// if err == nil {
		// 	CheckIngressSecurityGroup(ctx, d.Domain(), namespace.Namespace, d.DetectorInput.AwsConfig, &ingresses,
		// 		eksClusterInfo.Cluster.ResourcesVpcConfig.VpcId, &problems)
		// } else {
		// 	golog.Println("WARN - AWS configuration not provided, skip AWS resources validation.")
		// }

		checkSSLCertificate(ctx, d.Domain(), namespace.Namespace, d.DetectorInput.AwsConfig, &ingresses, &problems)
	}
	return problems, nil
}

func checkDynamicAnnotation(target string) bool {
	for _, annotation := range ValidDynamicAnnotationList {
		if strings.HasPrefix(target, fmt.Sprint(ALBAnnotationPrefix, "/", annotation)) {
			return true
		}
	}
	return false
}

// Compare the annotations against below URL, only checking annotations begin with alb.ingress.kubernetes.io
// https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/guide/ingress/annotations/
func checkAnnotation(domainName problem.DomainName, namespace string, ingressList *network.IngressList,
	ingresses *[]network.Ingress, problems *[]problem.Problem) {
	log.S().Info("Checking ingress annotations")
	affectedResources, sslAffectedResources := make(map[string]problem.ResourceDetails),
		make(map[string]problem.ResourceDetails)
	annotationProblem, sslAnnotationProblem := problem.Problem{}, problem.Problem{}
	for _, ingress := range ingressList.Items {
		hasValidSsl := false
		details := map[string]string{}
		for key, value := range ingress.Annotations {
			if strings.HasPrefix(key, ALBAnnotationPrefix) {
				isValid := false
				for _, annotation := range ValidAnnotationList {
					if fmt.Sprint(ALBAnnotationPrefix, "/", annotation) == key ||
						checkDynamicAnnotation(key) {
						isValid = true
						break
					}
				}
				if !isValid {
					details[key] = value
				}
			}
			if key == SSLAnnotation {
				hasValidSsl = true
			}
		}
		if len(details[InvalidAnnotationTitle]) > 0 {
			solution := map[string]interface{}{
				"ingressName": ingress.Name,
				"annotations": mapToString(details),
				"namespace":   namespace,
			}
			addIssueToDetails(affectedResources, ingress,
				details, InvalidAnnotationSolution, solution)

		}
		if !hasValidSsl {
			solution := map[string]interface{}{
				"ingressName":   ingress.Name,
				"sslAnnotation": SSLAnnotation,
			}
			addIssueToDetails(sslAffectedResources, ingress,
				details, SSLAnnotationSolution, solution)
		}
		*ingresses = append(*ingresses, ingress)
	}
	if len(affectedResources) > 0 {
		doc, err := url.Parse(SSLAnnotationDoc)
		if err != nil {
			log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
		}
		annotationProblem = problem.Problem{
			DomainName:        domainName,
			Name:              InvalidAnnotation,
			Description:       InvalidAnnotationDesc,
			Docs:              []*url.URL{doc},
			Tags:              InvalidIngressTags,
			Level:             problem.UserNamespace,
			AffectedResources: affectedResources,
		}
		*problems = append(*problems, annotationProblem)
	}
	if len(sslAffectedResources) > 0 {
		doc, err := url.Parse(SSLAnnotationDoc)
		if err != nil {
			log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
		}
		sslAnnotationProblem = problem.Problem{
			DomainName:        domainName,
			Name:              SSLAnnotation,
			Description:       SSLAnnotationDesc,
			Docs:              []*url.URL{doc},
			Tags:              InvalidIngressTags,
			Level:             problem.UserNamespace,
			AffectedResources: sslAffectedResources,
		}
		*problems = append(*problems, sslAnnotationProblem)
	}
}

// Compare the ingress rule paths with EKS service name and port.
func checkService(ctx context.Context, domainName problem.DomainName, namespace kubeclient.NamespacedName,
	client *kubeclient.KubeClient, ingresses *[]network.Ingress, problems *[]problem.Problem) {
	serviceList := &v1.ServiceList{}
	listOptions := metav1.ListOptions{}
	client.List(ctx, serviceList, namespace, listOptions)
	affectedResources := make(map[string]problem.ResourceDetails)
	var ingressPathProblem problem.Problem
	for _, ingress := range *ingresses {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				log.S().Infof("Checking ingress path %s, service %s, port %d", path.Path,
					path.Backend.Service.Name, path.Backend.Service.Port.Number)
				serviceExist := false
				for _, service := range serviceList.Items {
					if serviceExist {
						break
					}
					if service.Name == path.Backend.Service.Name {
						for _, port := range service.Spec.Ports {
							if port.Port == path.Backend.Service.Port.Number {
								serviceExist = true
								break
							}
						}
					}
				}
				if !serviceExist {
					log.S().Infof("Found ingress rule issue, service %s, port %d, path %s",
						path.Backend.Service.Name, path.Backend.Service.Port.Number, path.Path)
					details := map[string]string{
						"serviceName": path.Backend.Service.Name,
						"path":        path.Path,
					}
					if path.Backend.Service.Port.Number != 0 {
						details["portNumber"] = fmt.Sprint(path.Backend.Service.Port.Number)
					}
					solution := map[string]interface{}{
						"ingressName": ingress.Name,
						"rules":       mapToString(details),
						"namespace":   namespace.Namespace,
					}
					addIssueToDetails(affectedResources, ingress, details,
						InvalidServiceSolution, solution)
				}
			}
		}
		doc, err := url.Parse(InvalidIngressPathDoc)
		if err != nil {
			log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
		}
		if len(affectedResources[ingress.Name].Details) > 0 {
			item := affectedResources[ingress.Name]
			item.NextSteps = append(item.NextSteps, kubernetes.GetSolutionsByTemplate(InvalidServiceStep2Solution, map[string]interface{}{
				"ingressName": ingress.Name,
				"namespace":   namespace.Namespace,
			}, true)...)
			affectedResources[ingress.Name] = item
		}

		ingressPathProblem = problem.Problem{
			DomainName:        domainName,
			Name:              InvalidIngressPathTitle,
			Description:       InvalidIngressPathDesc,
			Docs:              []*url.URL{doc},
			Tags:              InvalidIngressTags,
			Level:             problem.UserNamespace,
			AffectedResources: affectedResources,
		}
	}
	if len(ingressPathProblem.AffectedResources) > 0 {
		*problems = append(*problems, ingressPathProblem)
	}
}

// Check AWS application load balancer, find the ALB by tag key ingress.k8s.aws/stack
func checkAlb(ctx context.Context, domainName problem.DomainName, input *problem.DetectorCreationInput,
	namespace string, ingresses *[]network.Ingress, problems *[]problem.Problem) {

	if input.AwsConfig.Region == "" {
		log.S().Warnf("AWS configuration not provided, skip AWS resources validation.")
		return
	}

	client := elb.NewFromConfig(input.AwsConfig)
	output, err := client.DescribeLoadBalancers(ctx, &elb.DescribeLoadBalancersInput{})

	if err != nil {
		log.S().Warnf("Error occurred while get AWS ALB info, error is %s", err)
		return
	}

	length := len(*ingresses)
	resChan := make(chan []problem.Problem, length)
	wg.Add(length)
	for _, ingress := range *ingresses {
		go checkIngressAlb(ctx, domainName, namespace, ingress, input, output, resChan)
	}
	wg.Wait()
	close(resChan)
	for val := range resChan {
		*problems = append(*problems, val...)
	}
}

func checkIngressAlb(ctx context.Context, domainName problem.DomainName, namespace string, ingress network.Ingress,
	input *problem.DetectorCreationInput, output *elb.DescribeLoadBalancersOutput,
	resChan chan []problem.Problem) {

	defer wg.Done()
	tagValue := fmt.Sprintf("%s/%s", namespace, ingress.Name)
	albExist := false
	problems := make([]problem.Problem, 0)

	var targetAlb types.LoadBalancer

	for _, ing := range ingress.Status.LoadBalancer.Ingress {
		albExist = false
		for _, alb := range output.LoadBalancers {
			if ing.Hostname == *alb.DNSName {
				albExist = true
				targetAlb = alb
				break
			}
		}
		if !albExist {
			break
		}
	}

	errorMsg := ""
	if albExist {
		switch targetAlb.State.Code {
		case types.LoadBalancerStateEnumActive:
			log.S().Infof("ALB %s is in %s state", tagValue, targetAlb.State.Code)
		case types.LoadBalancerStateEnumProvisioning:
			errorMsg = fmt.Sprintf("ALB %s is in %s state", tagValue, targetAlb.State.Code)
			log.S().Warn(errorMsg)
		case types.LoadBalancerStateEnumFailed:
			errorMsg = fmt.Sprintf("ALB %s is in %s state", tagValue, targetAlb.State.Code)
			log.S().Warn(errorMsg)
		case types.LoadBalancerStateEnumActiveImpaired:
			errorMsg = fmt.Sprintf("ALB %s is in %s state", tagValue, targetAlb.State.Code)
			log.S().Warn(errorMsg)
		default:
			errorMsg = fmt.Sprintf("ALB %s is in unknown state", tagValue)
			log.S().Warnf("ALB %s is in unknown state", tagValue)
		}
		if errorMsg != "" {
			problems = append(problems, *createProblem(domainName, &ingress, "ALB", tagValue, errorMsg))
		}
	} else {
		errorMsg = fmt.Sprintf("ALB %s not found", tagValue)
		log.S().Warn(errorMsg)
		doc, err := url.Parse(InvalidIngressPathDoc)
		if err != nil {
			log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
		}
		var deeplink *url.URL
		if input.LogRetriever != nil {
			endTime := time.Now()
			logsLink := input.LogDeeplinkRetriever.GetLogDeepLink(
				input.ClusterName, namespace, ingress.Name, true, true,
				kubernetes.SetStartTime(endTime, kubernetes.EventLogTimespan), endTime)
			deeplink, err = url.Parse(logsLink)
			if err != nil {
				log.S().Warnf("Log url generation failed %s", err)
			}
		}
		problem := problem.Problem{
			DomainName:  domainName,
			Name:        ALBNotExistTitle,
			Description: ALBNotExistDesc,
			Docs:        []*url.URL{doc},
			Tags:        InvalidIngressTags,
			Level:       problem.UserNamespace,
			AffectedResources: map[string]problem.ResourceDetails{
				ingress.Name: {
					Resource: &ingress,
					Details:  map[string]string{},
					Deeplink: map[problem.DeeplinkType]*url.URL{problem.DeeplinkKubeletLog: deeplink},
					NextSteps: kubernetes.GetSolutionsByTemplate(ALBNotCreatedSolution,
						map[string]interface{}{
							"alb": tagValue,
						}, true),
				},
			},
		}
		problems = append(problems, problem)
	}
	resChan <- problems
}

func createProblem(domainName problem.DomainName, ingress *network.Ingress, errorType string,
	key string, value string) *problem.Problem {
	doc, err := url.Parse(ALBIssueDoc)
	if err != nil {
		log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
	}
	return &problem.Problem{
		DomainName:  domainName,
		Name:        ALBIssueTitle,
		Description: ALBIssueDesc,
		Docs:        []*url.URL{doc},
		Tags:        InvalidIngressTags,
		Level:       problem.UserNamespace,
		AffectedResources: map[string]problem.ResourceDetails{
			ingress.Name: {
				Resource: ingress,
				Details: map[string]string{
					key: value,
				},
			},
		},
	}
}

// Quote from AWS document, "If a target group contains only unhealthy registered targets, the load balancer routes
// requests to all those targets, regardless of their health status.", if all targets are unhealthy, even if traffic
// will not be impacted. We will report the target health issue to user.
func checkTargetGroups(ctx context.Context, domainName problem.DomainName, namespace string, awsConfig aws.Config,
	ingresses *[]network.Ingress, problems *[]problem.Problem) {
	if awsConfig.Region == "" {
		log.S().Warn("AWS configuration not provided, skip AWS resources validation.")
		return
	}
	targetGroupTags := map[string]*tag.GetResourcesOutput{}
	targetGroupMap := map[string]types.TargetGroup{}

	client := elb.NewFromConfig(awsConfig)
	output, err := client.DescribeTargetGroups(ctx, &elb.DescribeTargetGroupsInput{})

	if err != nil {
		log.S().Warnf("Error occurred while get AWS target group info, error is %s", err)
		return
	}

	tagClient := tag.NewFromConfig(awsConfig)
	length := len(output.TargetGroups)
	tagChan := make(chan TargetGroupTags, length)
	wg.Add(length)
	for _, targetGroup := range output.TargetGroups {
		arn := targetGroup.TargetGroupArn
		targetGroupMap[*arn] = targetGroup
		go func(arn *string) {
			defer wg.Done()
			tagOutput, err := awsclient.GetTagByArn(ctx, tagClient, arn)
			if err != nil {
				log.S().Warnf("Error occurred while get AWS target group tag, error is %s", err)
			}
			tagChan <- TargetGroupTags{arn, tagOutput}
		}(arn)
	}
	wg.Wait()
	close(tagChan)
	for val := range tagChan {
		targetGroupTags[*val.arn] = val.tagOutput
	}

	resources := map[string]problem.ResourceDetails{}
	length = len(*ingresses)
	resChan := make(chan map[string]problem.ResourceDetails, length)
	wg.Add(length)
	for _, ingress := range *ingresses {
		go checkIngressTargetGroups(ctx, namespace, client, ingress, targetGroupMap, targetGroupTags, resChan)
	}
	wg.Wait()
	close(resChan)
	for val := range resChan {
		for k, v := range val {
			resources[k] = v
		}
	}

	doc, err := url.Parse(TargetGroupIssueDoc)
	if err != nil {
		log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
	}
	if len(resources) > 0 {
		problem := problem.Problem{
			DomainName:        domainName,
			Name:              TargetGroupIssueTitle,
			Description:       TargetGroupIssueDesc,
			Docs:              []*url.URL{doc},
			Tags:              InvalidIngressTags,
			Level:             problem.UserNamespace,
			AffectedResources: resources,
		}
		*problems = append(*problems, problem)
	}
}

func checkIngressTargetGroups(ctx context.Context, namespace string, client *elb.Client, ingress network.Ingress,
	targetGroutMap map[string]types.TargetGroup, targetGroupTag map[string]*tag.GetResourcesOutput,
	resChan chan map[string]problem.ResourceDetails) {
	defer wg.Done()
	resources := map[string]problem.ResourceDetails{}

	tagKey := "ingress.k8s.aws/stack"
	tagValue := fmt.Sprintf("%s/%s", namespace, ingress.Name)
	targetGroupExist := false

	var targetResource types.TargetGroup
	for k, targetGroup := range targetGroupTag {
		for _, res := range targetGroup.ResourceTagMappingList {
			for _, tag := range res.Tags {
				if *tag.Key == tagKey && *tag.Value == tagValue {
					targetGroupExist = true
					targetResource = targetGroutMap[k]
					break
				}
			}
		}
	}
	healthDetails := map[string]string{}
	if !targetGroupExist {
		log.S().Warn(fmt.Sprintf("Target group %s not found", tagValue))
	} else {
		targetGroupInput := &elb.DescribeTargetHealthInput{
			TargetGroupArn: targetResource.TargetGroupArn,
		}
		targetGroupResp, err := client.DescribeTargetHealth(ctx, targetGroupInput)
		if err == nil {
			for _, health := range targetGroupResp.TargetHealthDescriptions {
				if health.TargetHealth.State == types.TargetHealthStateEnumUnhealthy {
					// Add all target group issues to the problem's ingress resource.
					healthDetails[*health.Target.Id] = string(health.TargetHealth.Reason)
				}
			}
		} else {
			log.S().Warnf("Error occurred while get AWS target group health, error is %s", err)
		}
	}
	if len(healthDetails) > 0 {
		resources = map[string]problem.ResourceDetails{
			ingress.Name: {
				Resource: &ingress,
				Details:  healthDetails,
				NextSteps: kubernetes.GetSolutionsByTemplate(UnhealthTargetGroupSolution,
					map[string]interface{}{
						"targets": mapToString(healthDetails),
					}, true),
			},
		}
	}
	resChan <- resources
}

func GetEksClusterInfo(ctx context.Context, awsConfig aws.Config, clusterName string) (*eks.DescribeClusterOutput, error) {
	if awsConfig.Region == "" {
		log.S().Warn("AWS configuration not provided, skip AWS resources validation.")
		return nil, errors.New("FAILED TO GET AWS EKS CLUSTER INFO")
	}
	client := eks.NewFromConfig(awsConfig)
	params := eks.DescribeClusterInput{
		Name: &clusterName,
	}
	return client.DescribeCluster(ctx, &params)
}

// Check the security group, currently disabled on current release, maybe enabled in future.
func CheckIngressSecurityGroup(ctx context.Context, domainName problem.DomainName, namespace string, awsConfig aws.Config,
	ingresses *[]network.Ingress, vpcId *string, problems *[]problem.Problem) {
	if awsConfig.Region == "" {
		log.S().Warn("AWS configuration not provided, skip AWS resources validation.")
		return
	}
	client := ec2.NewFromConfig(awsConfig)

	for _, ingress := range *ingresses {
		securityGroup := ""
		for k, v := range ingress.Annotations {
			if k == "alb.ingress.kubernetes.io/security-groups" {
				securityGroup = v
				break
			}
		}
		if securityGroup == "" {
			log.S().Info("ALB security group not found")
			continue
		}

		// Masked arn in log, display directly in UI since there is authentication and authorization.
		log.S().Info("Found ALB security group")
		params := ec2.DescribeSecurityGroupsInput{}
		// The AWS ALB ingress security group supports both id and name.
		if strings.HasPrefix(securityGroup, "sg-") {
			params.GroupIds = []string{securityGroup}
		} else {
			groupName := "group-name"
			vpcIdName := "vpc-id"
			params.Filters = []ec2types.Filter{
				{
					Name:   &groupName,
					Values: []string{securityGroup},
				},
				{
					Name:   &vpcIdName,
					Values: []string{*vpcId},
				},
			}
		}
		output, err := client.DescribeSecurityGroups(ctx, &params)
		if err != nil {
			log.S().Warnf("Error occurred while get AWS security group tag, error is %s", err)
			break
		}
		// TODO need to check security group destination.
		if len(output.SecurityGroups) == 0 {
			log.S().Warn("No security group defined.")
		}
		// for _, sg := range output.SecurityGroups {
		// 	golog.Println("INFO - Checking security group for ALB")
		// }
	}
}

func checkSSLCertificate(ctx context.Context, domainName problem.DomainName, namespace string, awsConfig aws.Config,
	ingresses *[]network.Ingress, problems *[]problem.Problem) {
	if awsConfig.Region == "" {
		log.S().Warn("AWS configuration not provided, skip AWS resources validation.")
		return
	}

	log.S().Info("Checking ingress certificate")
	client := acm.NewFromConfig(awsConfig)

	affectedResources := make(map[string]problem.ResourceDetails)
	var sslProblem problem.Problem
	certHostnames := []string{}
	hasAwsError := false

	for _, ingress := range *ingresses {
		ingressHostnames := []string{}
		certificates := []string{}
		for key, value := range ingress.Annotations {
			if key == SSLAnnotation {
				certificates = strings.Split(value, SSLSeparator)
			}
		}
		for _, rule := range ingress.Spec.Rules {
			ingressHostnames = append(ingressHostnames, rule.Host)
		}
		if len(certificates) > 0 {
			for _, certArn := range certificates {
				req := acm.DescribeCertificateInput{
					CertificateArn: &certArn,
				}
				output, err := client.DescribeCertificate(ctx, &req)
				if err != nil {
					// If certificate arn does not exist, AWS sdk will throw error.
					if strings.Contains(err.Error(), "ResourceNotFoundException") {
						addIssueToDetails(affectedResources, ingress,
							map[string]string{"Certificate ARN not exist": certArn}, InvalidCertificateSolution,
							map[string]interface{}{
								"certificates": certArn,
							})
					} else {
						log.S().Warn("Error occurred while get AWS certificate")
						hasAwsError = true
					}
				} else {
					checkCertValidPeriod(output.Certificate.NotAfter, output.Certificate.NotBefore, certArn, affectedResources, ingress, time.Now())
					certHostnames = append(certHostnames, *output.Certificate.DomainName)
					for _, option := range output.Certificate.DomainValidationOptions {
						certHostnames = append(certHostnames, *option.DomainName)
					}
				}
			}
		} else {
			log.S().Warnf("Could not find SSL annotation %s in ingress %s", SSLAnnotation,
				ingress.Name)
		}
		if !hasAwsError {
			for _, ingressHostname := range ingressHostnames {
				if !validateIngressHost(ingressHostname, certHostnames) {
					addIssueToDetails(affectedResources, ingress,
						map[string]string{"Cannot find ingress host in certificate": ingressHostname},
						InvalidCertificateHostSolution, map[string]interface{}{
							"hostName":     ingressHostname,
							"certificates": strings.Join(certHostnames, "\n"),
						})
				}
			}
			doc, err := url.Parse(InvalidSSLCertDoc)
			if err != nil {
				log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
			}
			sslProblem = problem.Problem{
				DomainName:        domainName,
				Name:              InvalidSSLCertTitle,
				Description:       InvalidSSLCertDesc,
				Docs:              []*url.URL{doc},
				Tags:              InvalidIngressTags,
				Level:             problem.UserNamespace,
				AffectedResources: affectedResources,
			}
		}
	}
	if len(sslProblem.AffectedResources) > 0 {
		log.S().Warnf("Found issue found with ingress certificate")
		*problems = append(*problems, sslProblem)
	}
}

//Ingress hosts can be precise matches (for example “foo.bar.com”) or a wildcard (for example “*.foo.com”)
func validateIngressHost(ingressHostname string, certHostnames []string) bool {
	for _, hostname := range certHostnames {
		var result strings.Builder
		for i, literal := range strings.Split(hostname, "*") {
			if i > 0 {
				result.WriteString(".*")
			}
			result.WriteString(regexp.QuoteMeta(literal))
		}
		res, _ := regexp.MatchString(result.String(), ingressHostname)
		if res {
			return true
		}
	}
	return false
}

func checkCertValidPeriod(notAfter *time.Time, notBefore *time.Time, certArn string,
	affectedResources map[string]problem.ResourceDetails, ingress network.Ingress, currentTime time.Time) {
	maskedArn := kubernetes.MaskString(certArn)
	if !notAfter.After(currentTime) {
		log.S().Warnf("AWS certificate %s expired, certificate valid until %s",
			maskedArn, notAfter)
		addIssueToDetails(affectedResources, ingress,
			map[string]string{ExpiredSSLCertTitle + " " + certArn: fmt.Sprintf(
				"Certificate %s expired at %s", certArn, notAfter.String())},
			InvalidCertificateSolution, map[string]interface{}{
				"certificates": certArn,
			})
	} else if !notBefore.Before(currentTime) {
		log.S().Warnf("AWS certificate %s is not active, certificate will be valid after %s",
			maskedArn, notBefore)
		addIssueToDetails(affectedResources, ingress,
			map[string]string{
				InactiveSSLCertTitle + " " + certArn: fmt.Sprintf(
					"Certificate %s will be activate after %s", certArn, notBefore.String())},
			InvalidCertificateSolution, map[string]interface{}{
				"certificates": certArn,
			})
	}
}

func addIssueToDetails(resource map[string]problem.ResourceDetails, ingress network.Ingress,
	details map[string]string, solution string, templateValue map[string]interface{}) {
	if _, ok := resource[ingress.Name]; ok {
		// Go does not support reference the item in the map directly.
		// resource[ingress.Name].NextSteps = append(resource[ingress.Name].NextSteps, mapToString(details))
		detailsItem := resource[ingress.Name]
		detailsItem.NextSteps = append(detailsItem.NextSteps, mapToString(details))
		detailsItem.Details = details
		resource[ingress.Name] = detailsItem
	} else {
		resource[ingress.Name] = problem.ResourceDetails{
			Resource:  &ingress,
			Details:   details,
			NextSteps: kubernetes.GetSolutionsByTemplate(solution, templateValue, true),
		}
	}
}

func (d IngressEksDetector) Name() string {
	return IngressEksDetectorName
}

func (d IngressEksDetector) Domain() problem.DomainName {
	return problem.IngressFailuresDomain
}

// Convert the map to string, sort the order by keys, otherwise the output order will be inconsistent.
func mapToString(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("   %s: %s\n", k, m[k]))
	}
	return sb.String()
}

var ValidAnnotationList = []string{
	"load-balancer-name",
	"group.name",
	"group.order",
	"tags",
	"ip-address-type",
	"scheme",
	"subnets",
	"security-groups",
	"customer-owned-ipv4-pool",
	"load-balancer-attributes",
	"wafv2-acl-arn",
	"waf-acl-id",
	"shield-advanced-protection",
	"listen-ports",
	"ssl-redirect",
	"inbound-cidrs",
	"certificate-arn",
	"ssl-policy",
	"target-type",
	"backend-protocol",
	"backend-protocol-version",
	"target-group-attributes",
	"healthcheck-port",
	"healthcheck-protocol",
	"healthcheck-path",
	"healthcheck-interval-seconds",
	"healthcheck-timeout-seconds",
	"healthy-threshold-count",
	"unhealthy-threshold-count",
	"success-codes",
	"auth-type",
	"auth-idp-cognito",
	"auth-idp-oidc",
	"auth-on-unauthenticated-request",
	"auth-scope",
	"auth-session-cookie",
	"auth-session-timeout",
	"target-node-labels",
}

var ValidDynamicAnnotationList = []string{
	"actions.",
	"conditions.",
}
