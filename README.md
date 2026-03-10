# arxivfetcher

A simple tool to fetch and track ArXiv papers based on keywords. It saves paper abstracts to date-stamped directories and maintains a history to avoid duplicates.

## Features

- **Keyword Search**: Fetch papers matching a specific query.
- **Duplicate Prevention**: Tracks fetched paper IDs in a `history.txt` file.
- **Organized Storage**: Saves papers in `YYYYMMDD` subdirectories within your specified output folder.
- **Metadata Extraction**: Currently saves the title and abstract of each new paper.

## Prerequisites

- [Go](https://go.dev/doc/install) (v1.21 or later recommended)

## Setup

1. Clone the repository and navigate to the project directory.
2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Usage

### Building the Application

```bash
go build -o arxivfetcher main.go
```

### Running the Fetcher

You can run the binary with various flags:

```bash
./arxivfetcher --keyword "pigeonhole principle" --output_dir ./my_papers --max_results 20
```

#### Available Flags:

- `--keyword`: Search keyword for ArXiv (default: "string algorithm").
- `--output_dir`: Directory to store `history.txt` and fetched papers (default: `~/arxiv`).
- `--max_results`: Maximum number of results to fetch per request (default: 100).

### Example Output Structure

```text
my_papers/
├── history.txt
└── 20260310/
    ├── 2603.08558v1.txt
    └── 2603.08567v1.txt
```

### Scheduling with Cron

To automate fetching (e.g., daily at 9 AM):

```bash
0 9 * * * /path/to/arxivfetcher --keyword "machine learning" --output_dir /path/to/papers >> /path/to/fetcher.log 2>&1
```
