# mcp-ziti

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io) server that exposes the [OpenZiti](https://openziti.io) Management API to AI agents. It lets any MCP-compatible client — Claude Desktop, Cursor, VS Code Copilot, and others — create, inspect, and manage the resources in an OpenZiti network using natural language.

## How it works

The server communicates over STDIO. It can authenticate with an OpenZiti controller at startup using one of five methods (identity JSON file, username/password, client certificate, external JWT token, or OIDC client credentials), or it can start **without any credentials** and let the AI agent connect to a controller at runtime via the `connect-controller` tool.

## Quick start

1. [Download the binary](installation.md) for your platform
2. [Configure authentication](authentication.md) (or skip it and connect at runtime)
3. [Set up your AI agent](agent-setup.md) to use the server

## Prerequisites

- An [OpenZiti](https://openziti.io) controller reachable from the machine running the MCP server
