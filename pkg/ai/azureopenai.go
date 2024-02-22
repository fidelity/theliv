/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package ai

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/fidelity/theliv/pkg/config"
	rest "github.com/fidelity/theliv/pkg/http"
	log "github.com/fidelity/theliv/pkg/log"
	openai "github.com/sashabaranov/go-openai"
)

const (
	azureAIClientName = "azureopenai"
)

type AzureAIClient struct {
	nopCloser
	client      *openai.Client
	model       string
	temperature float32
}

type Token struct {
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	ExtExpiresIn int    `json:"ext_expires_in,omitempty"`
	Create       time.Time
}

var AzureConfig config.AzureConfig
var AiConfig config.AiConfig
var cachedToken Token

func (c *AzureAIClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	baseURL := config.GetBaseURL()
	engine := config.GetEngine()
	defaultConfig := openai.DefaultAzureConfig(token, baseURL)

	defaultConfig.AzureModelMapperFunc = func(model string) string {
		azureModelMapping := map[string]string{
			model: engine,
		}
		return azureModelMapping[model]
	}
	defaultConfig.APIType = openai.APIType(AiConfig.ApiType)
	defaultConfig.APIVersion = AiConfig.ApiVersion

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Azure OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	return nil
}

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to get openai response, error is %s.", err)
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *AzureAIClient) GetName() string {
	return azureAIClientName
}

func GetToken(ctx context.Context) (Token, error) {
	if cachedToken.AccessToken != "" &&
		time.Now().Before(
			cachedToken.Create.Add(time.Duration(cachedToken.ExpiresIn)*time.Second)) {
		return cachedToken, nil
	} else {
		return RefreshToken(ctx)
	}
}

func RefreshToken(ctx context.Context) (Token, error) {
	var token Token

	req := rest.NewRetryReq(10, 2, 10, 30)
	resp, err := req.
		SetFormData(map[string]string{
			"client_id":  AzureConfig.ClientID,
			"scope":      AzureConfig.Scope,
			"username":   AzureConfig.UserName,
			"password":   AzureConfig.Password,
			"grant_type": AzureConfig.GrantType,
		}).
		SetResult(&token).
		Post(AzureConfig.Endpoint)

	if err != nil || resp.StatusCode() != http.StatusOK {
		log.SWithContext(ctx).Errorf("Failed to get token, error is %s.", err)
		return cachedToken, err
	}
	token.Create = time.Now().Add(time.Second * 120)
	cachedToken = token
	return cachedToken, nil
}

func Init() {
	AiConfig = *config.GetThelivConfig().Ai
	AzureConfig = *config.GetThelivConfig().Azure
}
