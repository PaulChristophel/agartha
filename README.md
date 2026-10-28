[![Pull Request Quality Gate](https://github.com/PaulChristophel/sylvite/actions/workflows/ci.yaml/badge.svg)](https://github.com/PaulChristophel/sylvite/actions/workflows/ci.yaml)
[![Prepare and Publish Releases](https://github.com/PaulChristophel/sylvite/actions/workflows/release.yml/badge.svg?branch=master)](https://github.com/PaulChristophel/sylvite/actions/workflows/release.yml)

# Sylvite

*A Salt management console*

Sylvite is a web interface for [salt](https://github.com/saltstack/salt). The primary goal of this project is to provide a user-friendly interface for managing and interacting with job, event, minion, and return data, and offer a convenient interface to run salt commands.

## Features

- Dynamic Filtering: Search data dynamically.
- Detailed Job View: Expandable rows to view detailed job information, including job status, execution time, and result summary.
- Pagination: Efficient pagination to navigate through data.
- Minion Management: View and manage minions, including their status and assigned jobs.
- Salt Command Execution: Execute salt commands directly from the interface with a user-friendly input form.
- Authentication: Authenticate users through the configured local, LDAP, or CAS provider.
- Responsive Design: Fully responsive design, ensuring usability on various devices and screen sizes.

## Table of Contents

- Installation
- Usage
- Authorization
- Releases
- Contributing
- License

## Installation from source

To get started with Sylvite, follow these steps:

1. Clone the repository:

```bash
git clone https://github.com/PaulChristophel/sylvite.git
cd sylvite
```

2. Install dependencies:

```bash
make configure
```

3. Configure environment variables:

   Copy [config.example.yaml](./config.example.yaml) into the root directory of the application as `config.yaml` and update the necessary values.

4. Run the application:

```bash
ENV=debug make migrate
ENV=debug make run
```

## Validation

Run the backend tests and frontend validation before submitting changes:

```bash
env GOCACHE=/tmp/sylvite-gocache go test ./...
pnpm --dir web lint
pnpm --dir web build
```

## Database migrations

Sylvite now uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage database objects such as views, triggers, and materialized views. Running `make migrate` (or starting the server) automatically applies any pending SQL files stored under `server/db/migrations`. To add a new migration, create a matching pair of files following the `NNNN_description.up.sql` / `NNNN_description.down.sql` naming pattern in that directory and commit them alongside your feature.

## Usage

Once the application is running, open your web browser and navigate to http://localhost:8080 to access the Sylvite interface. Log in with your credentials and start managing your Salt jobs and minions.

## Authorization

Authenticated, active users can request a Salt eauth token. Salt authorizes commands from the user's LDAP group memberships and the generated `external_auth` configuration; the unused `is_staff` column is not an authorization boundary. Superusers continue to bypass LDAP group validation.

Local authentication validates passwords stored in `auth_user`; the former demonstration credentials are not accepted. LDAP identities are taken from the authenticated directory entry, and CAS identities are taken from a successful CAS service-validation assertion.

## Releases

Git tags are the source of truth for Sylvite release versions. Pull requests
run the quality gate and build the debug image variants with symbols retained.
Pull requests from branches in this repository also publish development images
with immutable version/date/commit tags and movable `pr-<number>` tags. Forked
pull requests build the same images without receiving registry credentials or
publishing them.

[Release Please](https://github.com/googleapis/release-please) maintains a
release pull request from conventional commit messages. A `fix:` commit proposes
a patch release, `feat:` proposes a minor release, and a commit marked as a
breaking change proposes a major release. Merging the release pull request
creates the semantic version tag and GitHub Release, then publishes the images
with that version.

The `.release-please-manifest.json` file tracks the most recent release version
and is maintained by Release Please. Development builds use that base version
and identify the exact source as
`<version>-<commit timestamp>-<commit>`. Release builds receive their clean
semantic version from the GitHub release tag.

The release workflow uses `RELEASE_PLEASE_TOKEN` when that secret is available
and otherwise falls back to the workflow token. Supplying a fine-grained token
allows GitHub to run the normal pull-request checks on bot-created release pull
requests.

## Contributing

We welcome contributions to Sylvite! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/your-feature-name`.
3. Make your changes.
4. Commit your changes.
5. Push the branch.
6. Open a pull request.

## License

Sylvite is licensed under the AGPL-3.0 License. See the [LICENSE](./LICENSE) file for more information.

## Renaming an existing installation

Sylvite was previously named Agartha. Update deployment references together:

- The executable is now `sylvite`; container application paths use `/opt/sylvite`.
- Configuration environment variables prefer the `SYLVITE_` prefix; legacy `AGARTHA_` variables remain supported. For the same setting, a nonempty `SYLVITE_` value takes precedence over `AGARTHA_`, and either prefix overrides file values. Empty environment values are ignored.
- Configuration files are searched in this order: `/etc/sylvite/`, `$HOME/.sylvite/`, the current directory, `/etc/agartha/`, and `$HOME/.agartha/`. The first matching file is used; files are not combined. An invalid file reports an error rather than falling back silently.
- The Salt authentication plugin is `extras/auth/sylvite.py` with eauth name `sylvite`. The self-contained legacy `extras/auth/agartha.py` remains available with eauth name `agartha`; existing Salt deployments can keep that filename and configuration. Set `salt.auth: agartha` (or `AGARTHA_SALT_AUTH=agartha`) when connecting to a master that still uses the legacy eauth name.
- Default database and user names are now `sylvite`. For an existing installation, explicitly configure the existing database name, user, and credentials to retain its data. Update Salt cache and returner connection settings consistently.
- Image and repository references in this project now use `sylvite`; publishing images and renaming the hosted repository are separate deployment operations.

The disposable migration test database remains named exactly `agartha_migration_test` to preserve its safety guard; its connection variable is `SYLVITE_TEST_DATABASE_URL`.
