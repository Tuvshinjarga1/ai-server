# AI Server

A Go-based AI server that integrates with OpenAI API and MCP (Model Context Protocol) for processing absence requests.

## Features

- OpenAI GPT-4 integration
- MCP (Model Context Protocol) support
- HTTP API endpoints
- Railway deployment ready

## API Endpoints

- `GET /` - Health check and server status
- `GET /health` - Health check endpoint
- `POST /process` - Process AI requests with OpenAI and MCP

## Environment Variables

- `OPENAI_API_KEY` - Your OpenAI API key
- `MCP_SERVER_URL` - URL of your MCP server (defaults to http://localhost:8080)
- `PORT` - Server port (set automatically by Railway)

## Deployment to Railway

1. Install Railway CLI:

   ```bash
   npm install -g @railway/cli
   ```

2. Login to Railway:

   ```bash
   railway login
   ```

3. Initialize Railway project:

   ```bash
   railway init
   ```

4. Set environment variables:

   ```bash
   railway variables set OPENAI_API_KEY=your_openai_api_key
   railway variables set MCP_SERVER_URL=your_mcp_server_url
   ```

5. Deploy:
   ```bash
   railway up
   ```

## Local Development

1. Set environment variables:

   ```bash
   export OPENAI_API_KEY=your_openai_api_key
   export MCP_SERVER_URL=http://localhost:8080
   ```

2. Run the server:

   ```bash
   go run main.go
   ```

3. Test the API:
   ```bash
   curl http://localhost:8080/
   curl -X POST http://localhost:8080/process
   ```

## Build

```bash
go build -o ai-server main.go
```
