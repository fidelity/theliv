/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package ingress

import (
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ActionType string

const (
	ActionTypeFixedResponse ActionType = "fixed-response"
	ActionTypeForward       ActionType = "forward"
	ActionTypeRedirect      ActionType = "redirect"
)

type Action struct {
	Type           ActionType `json:"type"`
	TargetGroupARN *string    `json:"targetGroupARN"`
	// +optional
	FixedResponseConfig *FixedResponseActionConfig `json:"fixedResponseConfig,omitempty"`
	// +optional
	RedirectConfig *RedirectActionConfig `json:"redirectConfig,omitempty"`
	// +optional
	ForwardConfig *ForwardActionConfig `json:"forwardConfig,omitempty"`
}

type FixedResponseActionConfig struct {
	// +optional
	ContentType *string `json:"contentType,omitempty"`
	// +optional
	MessageBody *string `json:"messageBody,omitempty"`

	StatusCode string `json:"statusCode"`
}

type RedirectActionConfig struct {
	// +optional
	Host *string `json:"host,omitempty"`
	// +optional
	Path *string `json:"path,omitempty"`
	// +optional
	Port *string `json:"port,omitempty"`
	// +optional
	Protocol *string `json:"protocol,omitempty"`
	// +optional
	Query *string `json:"query,omitempty"`

	StatusCode string `json:"statusCode"`
}

type ForwardActionConfig struct {
	TargetGroups []TargetGroupTuple `json:"targetGroups"`
	// +optional
	TargetGroupStickinessConfig *TargetGroupStickinessConfig `json:"targetGroupStickinessConfig,omitempty"`
}

type TargetGroupTuple struct {
	TargetGroupARN *string             `json:"targetGroupARN"`
	ServiceName    *string             `json:"serviceName"`
	ServicePort    *intstr.IntOrString `json:"servicePort"`
	// +optional
	Weight *int64 `json:"weight,omitempty"`
}

type TargetGroupStickinessConfig struct {
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
	// +optional
	DurationSeconds *int64 `json:"durationSeconds,omitempty"`
}
