
A simple MCP server that exposes a website fetching tool.

## Usage

Start the server using SSE transport:

# Using SSE transport on custom port
```bash
uv run mcp-simple-tool --transport sse --port 8000
```

The server exposes a tool named "fetch" that accepts one required argument:

- `url`: The URL of the website to fetch

## Example

Using the MCP client, run:

```
cd mcp_simple_tool && uv run client.py http://0.0.0.0:8000/sse
```
