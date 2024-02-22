/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	"github.com/fidelity/theliv/pkg/ai"
)

func Completion(ctx context.Context, prompt string) (interface{}, error) {
	token, err := ai.GetToken(ctx)
	if err != nil {
		return nil, err
	}
	aiConfig := GetAiConfig(token.AccessToken)
	var azClient = ai.AzureAIClient{}
	azClient.Configure(&aiConfig)
	return azClient.GetCompletion(ctx, prompt)
}

func GetAiConfig(token string) ai.AIProvider {
	return ai.AIProvider{
		Name:           ai.AiConfig.Name,
		Model:          ai.AiConfig.Model,
		Password:       token,
		BaseURL:        ai.AiConfig.BaseURL,
		EndpointName:   ai.AiConfig.EndpointName,
		Engine:         ai.AiConfig.Engine,
		Temperature:    ai.AiConfig.Temperature,
		ProviderRegion: ai.AiConfig.ProviderRegion,
		TopP:           ai.AiConfig.TopP,
		MaxTokens:      ai.AiConfig.MaxTokens,
	}
}
