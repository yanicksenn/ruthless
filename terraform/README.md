# Ruthless Terraform Deployment

This directory contains the Terraform configuration for deploying the Ruthless application to Google Cloud Platform. 
It uses a modular architecture with separate state for `preprod` and `prod` environments.

## Architecture

- **Artifact Registry**: Stores the frontend and backend Docker images.
- **Cloud SQL**: A managed PostgreSQL database for storage.
- **Cloud Run**: Serverless compute for running the frontend and backend containers.
- **Global HTTP(S) Load Balancer**: Provides routing and Google-managed SSL certificates for your custom domain.

## Deployment Instructions

### Prerequisites

1.  **Google Cloud CLI (`gcloud`)**: Installed and authenticated.
    ```bash
    gcloud auth login
    gcloud auth application-default login
    ```
2.  **Terraform**: Installed.
3.  **GCP Project**: You need a GCP project with billing enabled.

### Step-by-Step Deployment (for preprod)

1.  **Update variables:** 
    Open `environments/preprod/terraform.tfvars` and fill in your specific application values like the GCP `project_id`, your `domain`, OAuth credentials, and a secure `auth_secret`.

2.  **Build and Push Images:**
    Build your `backend` and `frontend` Docker images and push them to the GCP Artifact Registry (which Terraform will create for you in the next step). *Note: You may need to run terraform just to create the artifact registry first if the images aren't built yet, but assuming you build them when you deploy.*

3.  **Initialize Terraform:**
    ```bash
    cd environments/preprod
    terraform init
    ```

4.  **Review the Plan:**
    ```bash
    terraform plan
    ```

5.  **Apply Configuration:**
    ```bash
    terraform apply
    ```

6.  **Configure DNS:**
    After applying, Terraform will output a `load_balancer_ip`. Create an **A record** in your domain registrar pointing your `domain` (e.g., `preprod.ruthless.yanicksenn.com`) to this IP address.

    *Note: The Google-managed SSL certificate will stay in the "PROVISIONING" state until the DNS record propagates.*
