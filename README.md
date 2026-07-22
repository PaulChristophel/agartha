[![Pull Request Quality Gate](https://github.com/PaulChristophel/agartha/actions/workflows/ci.yaml/badge.svg)](https://github.com/PaulChristophel/agartha/actions/workflows/ci.yaml)
[![Prepare and Publish Releases](https://github.com/PaulChristophel/agartha/actions/workflows/release.yml/badge.svg?branch=master)](https://github.com/PaulChristophel/agartha/actions/workflows/release.yml)

# Agartha

Agartha is a web interface for [salt](https://github.com/saltstack/salt). The primary goal of this project is to provide a user-friendly interface for managing and interacting with job, event, minion, and return data, and offer a convenient interface to run salt commands.

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

To get started with Agartha, follow these steps:

1. Clone the repository:

```bash
git clone https://github.com/PaulChristophel/agartha.git
cd agartha
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
env GOCACHE=/tmp/agartha-gocache go test ./...
pnpm --dir web lint
pnpm --dir web build
```

## Database migrations

Agartha now uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage database objects such as views, triggers, and materialized views. Running `make migrate` (or starting the server) automatically applies any pending SQL files stored under `server/db/migrations`. To add a new migration, create a matching pair of files following the `NNNN_description.up.sql` / `NNNN_description.down.sql` naming pattern in that directory and commit them alongside your feature.

## Usage

Once the application is running, open your web browser and navigate to http://localhost:8080 to access the Agartha interface. Log in with your credentials and start managing your Salt jobs and minions.

## Authorization

Authenticated, active users can request a Salt eauth token. Salt authorizes commands from the user's LDAP group memberships and the generated `external_auth` configuration; the unused `is_staff` column is not an authorization boundary. Superusers continue to bypass LDAP group validation.

Local authentication validates passwords stored in `auth_user`; the former demonstration credentials are not accepted. LDAP identities are taken from the authenticated directory entry, and CAS identities are taken from a successful CAS service-validation assertion.

## Releases

Git tags are the source of truth for Agartha versions. Pull requests run the
quality gate and build every configured image without publishing anything.

[Release Please](https://github.com/googleapis/release-please) maintains a
release pull request from conventional commit messages. A `fix:` commit proposes
a patch release, `feat:` proposes a minor release, and a commit marked as a
breaking change proposes a major release. Merging the release pull request
creates the semantic version tag and GitHub Release, then publishes the images
with that version.

The `.release-please-manifest.json` file bootstraps the previous `0.11.3`
version and is maintained by Release Please. Builds do not read it. Development
builds use `0.0.0-dev.<commit>` and release builds receive their version from the
GitHub release tag.

The release workflow uses `RELEASE_PLEASE_TOKEN` when that secret is available
and otherwise falls back to the workflow token. Supplying a fine-grained token
allows GitHub to run the normal pull-request checks on bot-created release pull
requests.

## Contributing

We welcome contributions to Agartha! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/your-feature-name`.
3. Make your changes.
4. Commit your changes.
5. Push the branch.
6. Open a pull request.

## License

Agartha is licensed under the AGPL-3.0 License. See the [LICENSE](./LICENSE) file for more information.
