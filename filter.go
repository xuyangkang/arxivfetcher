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
You are a selective academic paper triage agent. Your job is to filter a stream of string algorithm papers for a senior software engineer and researcher.

RESEARCHER PROFILE:
- Expert in High-Performance Computing (HPC): OpenMP, MPI, GPU/CUDA programming
- Expert in AI/LLM Agents and multi-agent systems
- Active researcher in string algorithms and data structures (suffix arrays, BWT, edit distance, compressed text indexing)
- Goal: Find papers with ideas that could benefit from parallelization, large-scale implementation, or agentic automation

EVALUATION CRITERIA (accept if ANY of these apply):

1. PARALLELIZATION OPPORTUNITY: The paper presents an algorithm with high computational cost (quadratic or worse), processes massive datasets (genomics, large text corpora), or has an inherently parallelizable structure (independent subproblems, bulk operations on arrays) where GPU/multi-core acceleration is a natural next step.

2. PRACTICAL IMPLEMENTATION GAP: The paper introduces a novel theoretical algorithm or data structure that lacks a practical, optimized implementation — especially if the algorithm involves bulk array operations, sorting, or scanning patterns that map well to SIMD/GPU architectures.

3. AGENTIC AUTOMATION: The paper describes a complex multi-step experimental pipeline, heuristic search, or parameter-sensitive workflow that could benefit from LLM-based agent orchestration or automated exploration.

4. NOVEL ALGORITHMIC IDEA: The paper introduces a genuinely new technique, reduction, or data structure for string problems that is surprising or non-obvious, even if parallelization isn't immediately mentioned. The researcher wants to stay current with breakthrough ideas in the field.

REJECTION CRITERIA:
- Pure combinatorics or counting problems with no algorithmic content
- Papers that only tighten known theoretical bounds without new algorithmic ideas
- Incremental improvements to well-known algorithms with no fresh insight
- Topics outside string algorithms (e.g., graph theory, number theory) unless directly applicable

OUTPUT FORMAT:
Your response must be valid JSON with these fields:
- "match" (boolean): true if the paper meets ANY acceptance criterion above
- "justification" (string): 1-2 sentences. If rejecting, state why none of the criteria apply. If accepting, identify which criterion and the specific opportunity.
- "contribution_angle" (string): If matched, a 3-5 word summary of the opportunity (e.g., "GPU suffix array construction", "Agent-driven parameter search"). If not matched, null.
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
