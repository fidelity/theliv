/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package ai

import (
	"context"
)

type IAI interface {
	Configure(config IAIConfig) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	GetName() string
	Close()
}

type nopCloser struct{}

func (nopCloser) Close() {}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetEndpointName() string
	GetEngine() string
	GetTemperature() float32
	GetProviderRegion() string
	GetTopP() float32
	GetMaxTokens() int
}

type AIProvider struct {
	Name           string
	Model          string
	Password       string
	BaseURL        string
	EndpointName   string
	Engine         string
	Temperature    float32
	ProviderRegion string
	TopP           float32
	MaxTokens      int
}

func (p *AIProvider) GetBaseURL() string {
	return p.BaseURL
}

func (p *AIProvider) GetEndpointName() string {
	return p.EndpointName
}

func (p *AIProvider) GetTopP() float32 {
	return p.TopP
}

func (p *AIProvider) GetMaxTokens() int {
	return p.MaxTokens
}

func (p *AIProvider) GetPassword() string {
	return p.Password
}

func (p *AIProvider) GetModel() string {
	return p.Model
}

func (p *AIProvider) GetEngine() string {
	return p.Engine
}
func (p *AIProvider) GetTemperature() float32 {
	return p.Temperature
}

func (p *AIProvider) GetProviderRegion() string {
	return p.ProviderRegion
}
