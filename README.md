# arxivfetcher

A powerful, AI-driven CLI tool to fetch ArXiv papers based on keywords, rigorously filter them using the Grok foundation model, and directly download highly relevant PDFs into a unified research library.

## Features

- **Keyword Search**: Fetch papers matching complex Boolean queries natively via ArXiv's API.
- **AI-Powered Triage**: Feeds abstract previews directly into Grok to determine if a paper genuinely aligns with your research interests.
- **Automated PDF Downloads**: Extracts and downloads the full PDF of matches into your output directory.
- **Abstracted Storage Layer**: Decoupled backend architecture (`StorageBackend`) that currently supports local disk storage but is architected for future cloud (S3/GCS) integration.
- **Unified Local Saver**: Keeps everything in one place. Your `historyv2.txt` ledger and your matched PDFs live in the same directory for easy portability.
- **JSON-Lines State Engine (v2)**: Tracks paper progress (`related`, `unrelated`, `uploaded`) along with the **AI-generated justification** for each decision.
- **Rate Limiting**: Protect your API quotas with a built-in token-bucket rate limiter.

## Prerequisites

- [Go](https://go.dev/doc/install) (v1.26 or later recommended)
- A Grok API Key (OpenAI-compatible)

## Setup

1. Clone the repository and navigate to the project directory.
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Building the Application

```bash
go build -o arxivfetcher
```

### Running the Fetcher

```bash
export GROK_API_KEY="your-api-key"
./arxivfetcher --keyword "string algorithm" --max_results 20
```

#### Available Flags:

- `--keyword`: Search keyword for ArXiv (e.g., `'all:string AND cat:cs.DS'`).
- `--output_dir`: Unified directory for matched PDFs and the history ledger (default: `~/papers`).
- `--max_results`: Total maximum number of results to fetch in one run (default: 100).
- `--rps`: Rate limit in requests per second for AI calls and PDF downloads (default: 0.5).
- `--grok_apikey`: API key for Grok (can also be set via `GROK_API_KEY` env var).
- `--grok_baseurl`: Base URL for the OpenAI-compatible AI API (default: `https://api.x.ai/v1`).
- `--grok_model`: Model to use for the triage completion (default: `grok-4-1-fast-reasoning`).

### Advanced Queries

For better relevance, use ArXiv field prefixes and boolean operators:
- **Computer Science Algorithms**: `--keyword 'all:string AND cat:cs.DS'`
- **Specific Title Search**: `--keyword 'ti:"edit distance"'`

### Example Output Structure

```text
~/papers/
├── historyv2.txt      <-- JSON-lines state tracker with AI reasons
├── 2603.22591v1.pdf   <-- Matched PDF
├── 2603.11039v1.pdf   <-- Matched PDF
└── ...
```

### Scheduling with Cron

To automate fetching (e.g., every 6 hours):

```bash
0 0,6,12,18 * * * export GROK_API_KEY="your-key" && /path/to/arxivfetcher --keyword 'all:"suffix array" AND cat:cs.DS' >> /path/to/fetcher.log 2>&1
```
*(Note: Be mindful of your AI API quotas when scheduling automated searches).*
