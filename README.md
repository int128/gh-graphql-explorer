# gh-graphql-explorer [![go](https://github.com/int128/gh-graphql-explorer/actions/workflows/go.yaml/badge.svg)](https://github.com/int128/gh-graphql-explorer/actions/workflows/go.yaml)

This is a GitHub CLI extension that provides [GraphiQL](https://github.com/graphql/graphiql) explorer for GitHub GraphQL API.

## Getting Started

[GitHub CLI](https://cli.github.com) is required.
Install the extension:

```bash
gh extension install int128/gh-graphql-explorer
```

Run the extension:

```console
$ gh graphql-explorer
2026/01/24 14:26:19 gh-graphql-explorer is available at http://localhost:12345
```

Open the URL in your browser.
You can execute GraphQL queries with your GitHub authentication.

<img width="1402" height="821" alt="image" src="https://github.com/user-attachments/assets/6c6b4717-69ce-4109-8a22-b8cccff332b7" />

## How it works

This extension starts a local HTTP server that serves the following endpoints:

- `GET /`: Serves the GraphiQL page. See [index.html](index.html) for details.
- `POST /graphql`: Proxies the request to GitHub GraphQL API with authentication.
