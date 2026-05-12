# Postgres Operator

A Kubernetes operator that automates PostgreSQL database provisioning via a `PostgresDatabase` custom resource. Declare a database, user, and credentials in a CR and the operator creates them on your Postgres instance.

## How it works

The operator watches `PostgresDatabase` resources (API group `database.test.local/v1`). On each reconciliation it:

1. Connects to the Postgres admin instance via the `POSTGRES_ADMIN_URL` environment variable.
2. Checks whether the target database already exists.
3. If not, creates the database, user, password, and grants all privileges.

Provisioning is idempotent — if the database already exists the reconciler exits cleanly.

## Custom Resource

```yaml
apiVersion: database.test.local/v1
kind: PostgresDatabase
metadata:
  name: my-app-db
spec:
  database: my_app
  user: my_app_user
  password: s3cr3t           # plain-text; prefer passwordSecret in production
  passwordSecret:
    name: my-app-db-secret   # Kubernetes Secret containing the password
```

| Field            | Type   | Description                              |
|------------------|--------|------------------------------------------|
| `database`       | string | Name of the Postgres database to create  |
| `user`           | string | Postgres role/user to create             |
| `password`       | string | Password for the user                    |
| `passwordSecret` | object | Reference to a Secret holding the password |

## Prerequisites

- Go v1.24+
- Docker 17.03+
- kubectl v1.11.3+
- A running Kubernetes v1.11.3+ cluster
- A reachable PostgreSQL instance

## Getting Started

### Run locally (against the current kubeconfig cluster)

```sh
POSTGRES_ADMIN_URL="postgres://admin:pass@host:5432/postgres" make run
```

### Deploy to cluster

```sh
# Build and push the image
make docker-build docker-push IMG=<registry>/postgres-operator:tag

# Install CRDs
make install

# Deploy the controller (set POSTGRES_ADMIN_URL in config/manager/manager.yaml)
make deploy IMG=<registry>/postgres-operator:tag

# Apply a sample resource
kubectl apply -k config/samples/
```

### Uninstall

```sh
kubectl delete -k config/samples/
make uninstall
make undeploy
```

## Development

```sh
make manifests   # regenerate CRDs, RBAC, webhook configs
make generate    # regenerate DeepCopy methods
make fmt         # format code
make vet         # vet code
make test        # run unit tests
```

Run `make help` for the full list of targets.

## Project Distribution

### Single YAML bundle

```sh
make build-installer IMG=<registry>/postgres-operator:tag
kubectl apply -f dist/install.yaml
```

### Helm chart

```sh
kubebuilder edit --plugins=helm/v2-alpha
# chart generated under dist/chart/
```

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0. See [LICENSE](https://www.apache.org/licenses/LICENSE-2.0) for details.
