[![Go Report Card](https://goreportcard.com/badge/github.com/dispatchitinc/ecr-k8s-secret-creator)](https://goreportcard.com/report/github.com/dispatchitinc/ecr-k8s-secret-creator)

# ECR K8S Secret Creator

This application refreshes the ECR tokens that expire every 12 hours.  EKS has this capability built into the IAM roles, but when running outside of EKS you'll need to manage this functionality yourself with an instance profile.  There are many solutions for this like including a cron job and a third-party `aws-kubectl` image, but this solution worked best for us because:

- Many of the solutions out there use outdated manifests / apiVersions
- Uses a slim distroless container
- Deploy your own container (don't trust unvetted 3rd party images)
- Easier to manage logs
- Healthchecks
- Version lock your docker containers (not using latest)
- Uses roles/creds/etc. to generate the secret
- Helm chart for installation

This application creates a docker config.json (as a Kubernetes secret) that can authenticate docker clients to Amazon ECR. It is using the [ECR GetAuthorizationToken API](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_GetAuthorizationToken.html) to fetch the token from a specific AWS region.

*A special thanks to [bzon](https://github.com/bzon/ecr-k8s-secret-creator) for the initial thoughts.*
