# arxivfetcher

Yet another arxiv fetcher.

## Environment Setup

### 1. Install Go
Download and install the latest stable version of Go (v1.26.1):
```bash
curl -OL https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
rm go1.26.1.linux-amd64.tar.gz
```

### 2. Update PATH
Add Go to your PATH in `~/.bashrc`, `~/.zshrc`, or `~/.profile`:
```bash
# For bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# For zsh
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
source ~/.zshrc

# For universal (on most Linux systems)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
```

### 3. Initialize Project
```bash
go mod tidy
```

## Running the Application

To build the application:
```bash
go build -o arxivfetcher_bin main.go
```

To run the cron job:
```bash
./arxivfetcher_bin
```

To run once for debugging:
```bash
./arxivfetcher_bin --run_once
```

Alternatively, you can use `go run`:
```bash
go run main.go
```
