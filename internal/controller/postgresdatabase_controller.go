/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	databasev1 "github.com/j-selo/postgres-operator/api/v1"
	"github.com/j-selo/postgres-operator/postgres/postgres"
)

// PostgresDatabaseReconciler reconciles a PostgresDatabase object
type PostgresDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=database.test.local,resources=postgresdatabases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.test.local,resources=postgresdatabases/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=database.test.local,resources=postgresdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgresDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *PostgresDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Get the custom resources from the Kubernetes cluster
	var db databasev1.PostgresDatabase
	if err := r.Get(ctx, req.NamespacedName, &db); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Connect to the Postgres database
	pgClient, err := postgres.New(ctx, os.Getenv("POSTGRES_ADMIN_URL"))
	if err != nil {
		log.Error(err, "failed to connect to postgres")
		return ctrl.Result{}, err
	}
	defer pgClient.Close(ctx)

	// Check if database exists
	exists, err := pgClient.DatabaseExists(ctx, db.Spec.Database)
	if err != nil {
		log.Error(err, "failed to check database existence")
		return ctrl.Result{}, err
	}

	// Database exists
	if exists {
		log.Info("already provisioned", "database", db.Spec.Database)
		return ctrl.Result{}, nil
	}

	// Database provisioning
	if err := pgClient.Provision(ctx, db.Spec); err != nil {
		log.Error(err, "provision failed")
		return ctrl.Result{}, err
	}

	log.Info("provisioned", "database", db.Spec.Database, "user", db.Spec.User)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1.PostgresDatabase{}).
		Named("postgresdatabase").
		Complete(r)
}
