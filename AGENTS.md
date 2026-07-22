# AGENTS.md

## Tooling

- Use `pnpm`, never npm or yarn.
- Use Podman, never Docker.
- The project uses Go 1.26, Node.js 24, and pnpm 11.
- Preserve unrelated working-tree changes.

## Validation

For backend changes, run:

```bash
env GOCACHE=/tmp/agartha-gocache go test ./...
env GOCACHE=/tmp/agartha-gocache go vet ./...
```

For frontend changes, run:

```bash
pnpm --dir web lint
pnpm --dir web build
```

Run both backend and frontend validation for cross-cutting changes.

## Database Migrations

- Migration files belong in `server/db/migrations`.
- Every migration requires matching `NNNN_description.up.sql` and
  `NNNN_description.down.sql` files.
- Integration migration tests may reset the database.
- Only run them against a disposable PostgreSQL database named exactly
  `agartha_migration_test`.
- Never run migration integration tests against development or production data.

## Generated Files

- Do not manually edit generated Swagger files under `server/docs/v1`.
- Regenerate API documentation when API annotations change.

## Configuration and Secrets

- Use `config.example.yaml` as the documented configuration template.
- Never commit `config.yaml`, real credentials, Salt tokens, LDAP passwords,
  JWT secrets, or database connection strings.

## Destructive Commands

- Do not run `make clean`, `make podman-clean`, `podman volume prune`, or
  equivalent container or volume deletion commands without explicit approval.

## Change Scope

- Keep authentication and authorization changes separate from routine cleanup
  unless the task explicitly includes security redesign.
- Preserve existing API, Salt, database, and frontend behavior unless the task
  requests a contract change.

## Commits

- Use conventional commit prefixes such as `feat:`, `fix:`, `docs:`, and
  `chore:` because Release Please derives versions from commit messages.
