[![Test Release](https://github.com/PaulChristophel/agartha/actions/workflows/test.yaml/badge.svg?branch=master)](https://github.com/PaulChristophel/agartha/actions/workflows/test.yaml)
[![Production Release](https://github.com/PaulChristophel/agartha/actions/workflows/prod.yaml/badge.svg?branch=master)](https://github.com/PaulChristophel/agartha/actions/workflows/prod.yaml)

# Agartha

Agartha is a web interface for [salt](https://github.com/saltstack/salt). The primary goal of this project is to provide a user-friendly interface for managing and interacting with job, event, minion, and return data, and offer a convenient interface to run salt commands.

## Features

- Dynamic Filtering: Search data dynamically.
- Detailed Job View: Expandable rows to view detailed job information, including job status, execution time, and result summary.
- Pagination: Efficient pagination to navigate through data.
- Minion Management: View and manage minions, including their status and assigned jobs.
- Real-time Updates: Receive real-time updates on job status and results.
- Salt Command Execution: Execute salt commands directly from the interface with a user-friendly input form.
- Authentication & Authorization: Secure the interface with user authentication and role-based access control.
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
    Copy [config.example.yaml](./config.example.yaml) into the root directory of the application as `config.yaml` and update the necessary variables.

4. Run the application:
```bash
        ENV=debug make migrate
        ENV=debug make run
```

## Usage

Once the application is running, open your web browser and navigate to http://localhost:8080 to access the Agartha interface. Log in with your credentials and start managing your Salt jobs and minions.

## Contributing

We welcome contributions to Agartha! If you'd like to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch:
    git checkout -b feature/your-feature-name
3. Make your changes.
4. Commit your changes:
    git commit -m "Add your commit message"
5. Push to the branch:
    git push origin feature/your-feature-name
6. Open a pull request.

## License

Agartha is licensed under the AGPL-3.0 License. See the [LICENSE](./LICENSE) file for more information.
