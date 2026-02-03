package main

import (
	"context"
	"log"
	"time"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	// jsonschema "github.com/sashabaranov/go-openai/jsonschema"
)

// https://openai.com/api/pricing/
var GPT3_5_turbo string = "gpt-3.5-turbo-0125"
var GPT4_o string = "gpt-4o-2024-05-13"
var GPT4_turbo string = "gpt-4-turbo"
var GPT4_o_mini string = "gpt-4o-mini"
var GPT5_mini string = "gpt-5-mini"
var GPT5 string = "gpt-5"
var GPT5_2 string = "gpt-5.2"

type OpenAIRequest struct {
	prompt string
	model  string
	client *openai.Client
}

func retryWithBackoff(operation func() error, maxRetries int, verbose bool) error {
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			backoffTime := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s, 4s, 8s, 16s
			if verbose {
				log.Printf("[RETRY] Attempt %d/%d failed: %v. Retrying in %v...", attempt, maxRetries, err, backoffTime)
			}
			time.Sleep(backoffTime)
		}
	}
	return fmt.Errorf("failed after %d attempts: %v", maxRetries, err)
}

func fetchOpenAIAnswer(req OpenAIRequest, verbose bool) (string, error) {
	var result string

	err := retryWithBackoff(func() error {
		resp, err := req.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: req.model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: req.prompt,
					},
				},
			},
		)

		if err != nil {
			return err
		}

		result = resp.Choices[0].Message.Content
		return nil
	}, 5, verbose)

	if err != nil {
		if verbose {
			log.Printf("ChatCompletion error: %v\n", err)
		}
		return "", err
	}

	return result, nil
}

func fetchOpenAIAnswerJSON(req OpenAIRequest, schema openai.ChatCompletionResponseFormatJSONSchema, verbose bool) (string, error) {
	if verbose {
		log.Printf("[OPENAI] Making JSON request with model: %s", req.model)
		log.Printf("[OPENAI] Prompt length: %d characters", len(req.prompt))
	}

	var result string

	err := retryWithBackoff(func() error {
		resp, err := req.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: req.model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: req.prompt,
					},
				},
				ResponseFormat: &openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
					JSONSchema: &schema,
				},
			},
		)

		if err != nil {
			return err
		}

		result = resp.Choices[0].Message.Content
		return nil
	}, 5, verbose)

	if err != nil {
		if verbose {
			log.Printf("[OPENAI] ChatCompletion error: %v\n", err)
		}
		return "", err
	}

	if verbose {
		log.Printf("[OPENAI] ChatCompletion successful")
		log.Printf("[OPENAI] Response content length: %d characters", len(result))
	}
	return result, nil
}

