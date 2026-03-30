package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type PaperFilterResponse struct {
	Match             bool    `json:"match"`
	Justification     string  `json:"justification"`
	ContributionAngle *string `json:"contribution_angle"`
}

func filterPaper(ctx context.Context, apiKey, baseURL, model, paperText string) (*PaperFilterResponse, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for AI filtering")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	systemPrompt := `
You are a ruthless, highly selective academic triage agent. Your job is to filter a stack of academic papers (abstracts/summaries) for a researcher.

The researcher is relatively new to String Algorithms and Data Structures, but they are an expert in Parallel Computing (OpenMP, MPI, GPU) and Multi-Agent Systems.

Your goal is to reject as many papers as possible. You must ONLY flag a paper as a match if there is a glaring, obvious opportunity for the researcher to apply their specific engineering skills to improve the paper's work.
Strict Evaluation Rules:

    Default to REJECT (match: false). Do not use imagination to force a fit. If the connection isn't obvious, reject it.

    Look for Implementation Gap: ACCEPT the paper if it introduces a novel sequential algorithm, mentions high computational costs, or deals with massive datasets (e.g., genomics, huge text corpora) where GPU/Multi-core acceleration is a highly logical next step.

    Look for Automation: ACCEPT the paper if it describes a complex, multi-step experimental workflow or heuristic search that could clearly be automated or optimized using AI Agents.

Output Format:

Your response must be in valid JSON format with the following fields:

    "match" (boolean): true ONLY if the paper survives the strict evaluation rules above. Otherwise, false.

    "justification" (string): A ruthless, 1-2 sentence explanation. If rejecting, state exactly why their skills don't apply. If accepting, state the exact bottleneck or workflow they should target.

    "contribution_angle" (string): If matched, write a 3-5 word summary of the exact technical approach (e.g., "GPU parallelization of substring search", "Agentic workflow for hyperparameter tuning"). If not matched, output null.
`

	// Define JSON schema for structured output
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"match":              {Type: jsonschema.Boolean, Description: "Whether the paper matches the research interests"},
			"justification":      {Type: jsonschema.String, Description: "A brief justification for the match or non-match"},
			"contribution_angle": {Type: jsonschema.String, Description: "3-5 word summary of the technical approach if matched, else null"},
		},
		Required: []string{"match", "justification", "contribution_angle"},
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: paperText,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        "PaperFilterResponse",
				Description: "Structured response indicating whether a paper matches research interests",
				Schema:      &schema,
				Strict:      true,
			},
		},
		Temperature: 0.1,
	})

	if err != nil {
		return nil, fmt.Errorf("error calling Grok API: %v", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from API")
	}

	content := resp.Choices[0].Message.Content

	var filterResp PaperFilterResponse
	err = json.Unmarshal([]byte(content), &filterResp)
	if err != nil {
		log.Printf("Raw response: %s", content)
		return nil, fmt.Errorf("error parsing JSON response: %v", err)
	}

	return &filterResp, nil
}
