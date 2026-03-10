# arxivfetcher

A simple tool to fetch and track ArXiv papers based on keywords. It saves paper abstracts to directories named after the paper's **publication date** and maintains a history to avoid duplicates.

## Features

- **Keyword Search**: Fetch papers matching a specific query.
- **Duplicate Prevention**: Tracks fetched paper IDs in a `history.txt` file.
- **Organized Storage**: Saves papers in `YYYYMMDD` subdirectories based on their publication date.
- **Metadata Extraction**: Saves the title and abstract of each new paper.

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

- `--keyword`: Search keyword for ArXiv (e.g., `'all:string AND cat:cs.DS'`).
- `--output_dir`: Directory to store `history.txt` and fetched papers (default: `~/arxiv`).
- `--max_results`: Total maximum number of results to fetch in one run (default: 100).

### Advanced Queries

For better relevance, use ArXiv field prefixes and boolean operators:
- **Computer Science Algorithms**: `--keyword 'all:string AND cat:cs.DS'`
- **Specific Title Search**: `--keyword 'ti:"edit distance"'`

### Example Output Structure

```text
my_papers/
├── history.txt
└── 20260219/
    ├── 2602.17201v1.txt
    └── 2602.17202v1.txt
```

### Scheduling with Cron

To automate fetching (e.g., every hour at minute 15):

```bash
15 * * * * /path/to/arxivfetcher --keyword "string algorithm" --output_dir /path/to/papers >> /path/to/fetcher.log 2>&1
```
