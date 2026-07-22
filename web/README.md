# Agartha web application

This directory contains Agartha's React and Vite frontend. The production build is embedded in the Go server.

## Development

Install dependencies from the repository root:

```bash
pnpm --dir web install
```

Run the frontend development server:

```bash
pnpm --dir web dev
```

Vite proxies Agartha API, authentication, version, and documentation requests to the backend at `http://localhost:8080`.

## Validation

```bash
pnpm --dir web lint
pnpm --dir web build
```

The build output is written to `web/dist` and embedded by `main.go`.
