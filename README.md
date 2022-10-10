[![Go Report Card](https://goreportcard.com/badge/github.com/dispatchitinc/ecr-k8s-secret-creator)](https://goreportcard.com/report/github.com/dispatchitinc/ecr-k8s-secret-creator)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/4613f935eff94c6f860bd8409554331f)](https://www.codacy.com/gh/dispatchitinc/ecr-k8s-secret-creator/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=dispatchitinc/ecr-k8s-secret-creator&amp;utm_campaign=Badge_Grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/4613f935eff94c6f860bd8409554331f)](https://www.codacy.com/gh/dispatchitinc/ecr-k8s-secret-creator/dashboard?utm_source=github.com&utm_medium=referral&utm_content=dispatchitinc/ecr-k8s-secret-creator&utm_campaign=Badge_Coverage)

# ECR K8S Secret Creator

This application refreshes the ECR tokens that expire every 12 hours.  EKS has this capability built into the IAM roles, but when running outside of EKS you'll need to manage this functionality yourself with an instance profile.  There are many solutions for this like including a cron job and a third-party `aws-kubectl` image, but this solution worked best for us because:

- Uses a distroless container
- Deploy your own container (don't trust unvetted 3rd party images)
- Easier to manage logs
- Version lock your docker containers (not using latest)
- Uses roles/creds/etc. to generate the secret
- Helm chart for installation
- Deploy the secret to multiple namespaces
- SBOM generated for every version

This application creates a docker config.json (as a Kubernetes secret) that can authenticate docker clients to Amazon ECR. It is using the [ECR GetAuthorizationToken API](https://docs.aws.amazon.com/AmazonECR/latest/APIReference/API_GetAuthorizationToken.html) to fetch the token from a specific AWS region.

## Installation

You can add the helm chart below with a custom `values.yaml` file to override your specific settings.

```
helm repo add dispatch-secret-creator https://dispatchitinc.github.io/ecr-k8s-secret-creator/
helm install ecr-k8s-secret-creator dispatch-secret-creator/ecr-k8s-secret-creator
```

### IAM Policy

You will need to set up an IAM policy that is attached to an AWS EC2 IAM Role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "ecr:GetAuthorizationToken",
      "Resource": "*"
    }
  ]
}
```

```sh
aws iam create-policy \
  --policy-name ${GET_ECR_AUTH_IAM_POLICY} \
  --policy-document file://iam-policy.json \
  --description "A policy that can get ECR authorization token"
```

### IAM Role

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${AWS_PROFILE}:role/${IAM_K8S_NODE_ROLE}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

```sh
aws iam create-role \
  --role-name ${GET_ECR_AUTH_IAM_ROLE} \
  --assume-role-policy-document file://ec2-iam-trust.json
```

### Attaching the Policy

```sh
aws iam attach-role-policy \
  --role-name ${GET_ECR_AUTH_IAM_ROLE} \
  --policy-arn arn:aws:iam::${AWS_PROFILE}:policy/${GET_ECR_AUTH_IAM_POLICY}
```

## Contributors

A special thanks to @bzon for the initial thoughts.
