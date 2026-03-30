# arxivfetcher

A powerful, AI-driven CLI tool to fetch ArXiv papers based on keywords, rigorously filter them using the Grok foundation model, and directly download highly relevant PDFs for local consumption.

## Features

- **Keyword Search**: Fetch papers matching complex Boolean queries natively via ArXiv's API.
- **AI-Powered Triage**: Feeds abstract previews directly into Grok to determine if a paper genuinely aligns with your specific engineering/research skills (e.g. Parallel Computing vs purely theoretical data structures), rejecting anything that isn't a strong match.
- **Automated PDF Downloads**: Bypasses manual abstraction reading; automatically extracts and downloads the full PDF of any paper flagged as `Matched`.
- **JSON-Lines State Engine (v2)**: Tracks the pipeline phase of every paper utilizing a robust, crash-immune state machine recorded in `historyv2.txt` (`new` -> `unrelated` / `related` -> `uploaded`).
- **Flexible Storage**: Saves parsed PDFs straight to a configurable flattened directory (`~/papers`), discarding the bloat of abstract text files and subfolders. Extensible via the generic `PaperSaver` interface.

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

You can run the binary with various flags. By default, it expects a Grok API key via the terminal environment.

```bash
export GROK_API_KEY="your-api-key"
./arxivfetcher --keyword "string algorithm" --max_results 20
```

#### Available Flags:

- `--keyword`: Search keyword for ArXiv (e.g., `'all:string AND cat:cs.DS'`).
- `--output_dir`: Directory to store the `historyv2.txt` tracking ledger (default: `~/arxiv`).
- `--max_results`: Total maximum number of results to fetch in one run (default: 100).
- `--papers_dir`: Local directory to store successfully matched PDFs (default: `~/papers`).
- `--grok_apikey`: API key for Grok (can also be set via `GROK_API_KEY` environment variable).
- `--grok_baseurl`: Base URL for the OpenAI-compatible AI API (default: `https://api.x.ai/v1`).
- `--grok_model`: Model to use for the triage completion (default: `grok-4-1-fast-reasoning`).

### Advanced Queries

For better relevance, use ArXiv field prefixes and boolean operators:
- **Computer Science Algorithms**: `--keyword 'all:string AND cat:cs.DS'`
- **Specific Title Search**: `--keyword 'ti:"edit distance"'`

### Example Output Structure

```text
~/arxiv/
└── historyv2.txt

~/papers/
├── 2603.26176v1.pdf
├── 2603.22591v1.pdf
└── ...
```

### Scheduling with Cron

To automate fetching (e.g., every 6 hours):

```bash
0 0,6,12,18 * * * export GROK_API_KEY="your-key" && /path/to/arxivfetcher --keyword 'all:"suffix array" AND cat:cs.DS' >> /path/to/fetcher.log 2>&1
```
*(Note: Be mindful of your AI API quotas when scheduling heavily automated wide-net searches).*
