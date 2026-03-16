# Project: sh3r4rd.com Monorepo

## Project Overview

This is a monorepo for the personal website `sh3r4rd.com`. It consists of a React-based frontend and a serverless backend for a "Recruiter Dashboard".

### Key Technologies

*   **Frontend**:
    *   Framework: React
    *   Build Tool: Vite
    *   Styling: Tailwind CSS
*   **Backend (Recruiter Dashboard)**:
    *   Language: Go
    *   Platform: AWS Lambda
    *   Services: Amazon S3 (for raw email storage), DynamoDB (for parsed data), SES (for email receiving), and API Gateway (for a REST API).
*   **Infrastructure as Code**:
    *   Tool: Terraform
*   **CI/CD**:
    *   Platform: GitHub Actions for continuous deployment of the frontend.

### Directory Structure

*   `src/`: Contains the React frontend source code.
*   `infra/recruiter-dashboard/`: Contains the Terraform configuration for the backend infrastructure.
*   `infra/recruiter-dashboard/lambda-src/`: Contains the Go source code for the AWS Lambda functions.
*   `.github/workflows/`: Contains the GitHub Actions workflow for deployment.

## Building and Running

### Frontend

1.  **Install Dependencies**:
    ```bash
    npm install
    ```
2.  **Run Development Server**:
    ```bash
    npm run dev
    ```
3.  **Build for Production**:
    ```bash
    npm run build
    ```
    The output will be in the `dist/` directory.

4.  **Linting**:
    ```bash
    npm run lint
    ```

### Backend (Terraform)

The backend infrastructure is managed by Terraform.

1.  **Initialize Terraform**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard init
    ```
2.  **Plan Changes**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard plan
    ```
3.  **Apply Changes**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard apply
    ```
    *Note: You will need to create a `terraform.tfvars` file based on the `terraform.tfvars.example`.*

## Deployment

### Frontend

Deployment of the frontend to Amazon S3 is automated via a GitHub Actions workflow. The workflow is triggered on every push to the `main` branch.

You can also deploy manually using the Makefile:
```bash
make deploy bucket=your-s3-bucket-name
```

## Development Conventions

*   **Conventional Commits**: This repository follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
*   **Modular Terraform**: The Terraform code is organized into reusable modules located in `infra/recruiter-dashboard/modules/`.
*   **Go Backend Structure**: The Go Lambda functions follow a structured layout, separating handlers, database logic, and other concerns into internal packages.
