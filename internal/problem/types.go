/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

type ReportCard struct {
	Name            string                `json:"name"`
	RootCause       *ReportCardIssue      `json:"rootCause"`
	Resources       []*ReportCardResource `json:"resources"`
	TopResourceType string                `json:"topResourceType"`
	Level           ProblemLevel          `json:"level"`
	ID              string                `json:"id"`
}

type ReportCardIssue struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Solutions   []string          `json:"solutions,omitempty"`
	CreatedTime string            `json:"createdTime,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	DomainName  DomainName        `json:"domainName,omitempty"`
	// Documents   []string   `json:"documents,omitempty"`
	CauseLevel int `json:"causelevel,omitempty"`
}

type ReportCardResource struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Issue       *ReportCardIssue       `json:"issue,omitempty"`
	Deeplink    map[string]string      `json:"deeplink,omitempty"`
}

type helmChart struct {
	instance string
	version  string
	chart    string
	release  string
}

type ArgoInstance struct {
	Instance        string
	RolloutTemplate string
}
