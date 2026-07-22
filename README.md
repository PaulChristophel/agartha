[![Test Release](https://github.com/PaulChristophel/agartha/actions/workflows/test.yaml/badge.svg?branch=master)](https://github.com/PaulChristophel/agartha/actions/workflows/test.yaml)
[![Production Release](https://github.com/PaulChristophel/agartha/actions/workflows/prod.yaml/badge.svg?branch=master)](https://github.com/PaulChristophel/agartha/actions/workflows/prod.yaml)

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
