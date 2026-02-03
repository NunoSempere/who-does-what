package main

import (
	"context"
	"log"

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

type OpenAIRequest struct {
	prompt string
	model  string
	token  string
}

func fetchOpenAIAnswer(req OpenAIRequest) (string, error) {

	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: req.model, // openai.GPT4TurboPreview, // openai.GPT3Dot5Turbo // "gpt-3.5-turbo-0125"
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: req.prompt,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	result := resp.Choices[0].Message.Content
	return result, nil

}

func fetchOpenAIAnswerJSON(req OpenAIRequest, schema openai.ChatCompletionResponseFormatJSONSchema) (string, error) {
	log.Printf("[OPENAI] Creating OpenAI client and making JSON request with model: %s", req.model)
	log.Printf("[OPENAI] Prompt length: %d characters", len(req.prompt))
	
	client := openai.NewClient(req.token)
	resp, err := client.CreateChatCompletion(
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
		log.Printf("[OPENAI] ChatCompletion error: %v\n", err)
		return "", err
	}

	log.Printf("[OPENAI] ChatCompletion successful")
	result := resp.Choices[0].Message.Content
	log.Printf("[OPENAI] Response content length: %d characters", len(result))
	return result, nil
}

