# Yokai-Finder MCP

A Model Context Protocol (MCP) server for searching Japanese yokai (folklore creatures) books in the National Diet Library.

## Features

- **Search Yokai Books**: Search for books about yokai using the National Diet Library API
- **Flexible Search**: Search by yokai name, region, or category
- **Caching**: Built-in caching to reduce API calls and improve performance
- **MCP Protocol**: Fully compliant with the Model Context Protocol

## Installation

### Prerequisites

- Go 1.22 or higher
- Git

### Build from Source

```bash
git clone https://github.com/yourname/yokai-finder-mcp.git
cd yokai-finder-mcp
go build -o yokai-finder-mcp cmd/server/server.go
```

## Usage

### As an MCP Server

The server can be used as an MCP server by running it directly:

```bash
./yokai-finder-mcp
```

The server communicates via stdin/stdout using JSON-RPC protocol.

### Configuration for Claude Desktop

Add the following to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "yokai-finder": {
      "command": "go",
      "args": ["run", "cmd/server/server.go"],
      "cwd": "/path/to/yokai-finder-mcp",
      "env": {}
    }
  }
}
```

Or use the built binary:

```json
{
  "mcpServers": {
    "yokai-finder": {
      "command": "/path/to/yokai-finder-mcp/yokai-finder-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

## Available Tools

### search_yokai_books

Search for books about yokai in the National Diet Library.

**Parameters:**
- `name` (optional): Name of the yokai to search for (e.g., '河童', '天狗', '九尾の狐')
- `region` (optional): Region or prefecture associated with the yokai (e.g., '岩手', '京都')
- `category` (optional): Category of yokai (e.g., '水妖', '山妖', '動物妖怪')
- `limit` (optional): Maximum number of results to return (default: 10, max: 100)

**Example:**
```json
{
  "name": "search_yokai_books",
  "arguments": {
    "name": "河童",
    "limit": 5
  }
}
```

## Development

### Project Structure

```
yokai-finder-mcp/
├── cmd/
│   └── server/
│       └── server.go          # Main server entry point
├── internal/
│   ├── cache/
│   │   └── cache.go          # Caching functionality
│   ├── handler/
│   │   └── handler.go        # MCP request handlers
│   └── ndl/
│       └── ndl.go            # National Diet Library API client
├── pkg/
│   └── types/
│       └── types.go          # Type definitions
├── go.mod
├── mcp.json                  # MCP configuration
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o yokai-finder-mcp cmd/server/server.go
```

## API

The server implements the MCP protocol with the following methods:

- `initialize`: Initialize the MCP connection
- `tools/list`: List available tools
- `tools/call`: Execute a tool

## Caching

The server includes a built-in cache with the following features:
- TTL: 30 minutes
- Max size: 100 entries
- Automatic cleanup of expired entries

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- National Diet Library for providing the API
- Model Context Protocol for the protocol specification