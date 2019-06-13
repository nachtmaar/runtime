/*
Copyright 2019 The Kyma Authors.

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

package function

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	runtimev1alpha1 "github.com/kyma-incubator/runtime/pkg/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"crypto/sha256"
	runtimeUtil "github.com/kyma-incubator/runtime/pkg/utils"
)

var log = logf.Log.WithName("function_controller")

// Add creates a new Function Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFunction{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("function-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Function
	err = c.Watch(&source.Kind{Type: &runtimev1alpha1.Function{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Function - change this for objects you create
	// err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
	// 	IsController: true,
	// 	OwnerType:    &runtimev1alpha1.Function{},
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

var (
	fnConfigName                                   = "fn-config"
	fnConfigNamespace                              = "default"
	fnConfigNameEnv                                = os.Getenv("CONTROLLER_CONFIGMAP")
	fnConfigNamespaceEnv                           = os.Getenv("CONTROLLER_CONFIGMAP_NS")
	buildTemplateName                              = "function-kaniko"
	buildTemplateNamespace                         = "default"
	buildTemplateNameEnv                           = os.Getenv("BUILD_TEMPLATE")
	buildTemplateNamespaceEnv                      = os.Getenv("BUILD_TEMPLATE_NS")
	_                         reconcile.Reconciler = &ReconcileFunction{}
)

// ReconcileFunction is the controller.Reconciler implementation for Function objects
// ReconcileFunction reconciles a Function object
type ReconcileFunction struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="runtime.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="runtime.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list
// +kubebuilder:rbac:groups=";apps;extensions",resources=deployments,verbs=create;get;delete;list;update;patch
func (r *ReconcileFunction) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Get Function Controller Configuration
	fnConfig := &corev1.ConfigMap{}
	if err := r.getFunctionControllerConfiguration(fnConfig); err != nil {
		return reconcile.Result{}, err
	}

	// Get a new *RuntimeInfo
	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		log.Error(err, "Error while try to get a new RuntimeInfo instance", "namespace", fnConfig.Namespace, "name", fnConfig.Name)
		return reconcile.Result{}, err
	}

	// Get Function instance
	fn := &runtimev1alpha1.Function{}
	if err := r.getFunctionInstance(request, fn); err != nil {
		log.Error(err, "Error reading Function instance", "namespace", request.Namespace, "name", request.Name)
		return reconcile.Result{}, err
	}

	// Create Function's ConfigMap
	foundCm := &corev1.ConfigMap{}
	deployCm := &corev1.ConfigMap{}
	if err := r.createFunctionConfigMap(foundCm, deployCm, fn); err != nil {
		log.Error(err, "Error while try to create the Function's ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		return reconcile.Result{}, err
	}

	// Update Function's ConfigMap
	if err := r.updateFunctionConfigMap(foundCm, deployCm); err != nil {
		log.Error(err, "Error while try to upadate Function's ConfigMap:", "namespace", deployCm.Namespace, "name", deployCm.Name)
		return reconcile.Result{}, err
	}

	hash := sha256.New()
	hash.Write([]byte(foundCm.Data["handler.js"] + foundCm.Data["package.json"]))
	functionSha := fmt.Sprintf("%x", hash.Sum(nil))
	imageName := fmt.Sprintf("%s/%s-%s:%s", rnInfo.RegistryInfo, fn.Namespace, fn.Name, functionSha)

	if err := r.getFunctionBuildTemplate(rnInfo, fnConfig, fn, imageName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.buildFunctionImage(rnInfo, fnConfig, fn, imageName); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.serveFunction(rnInfo, foundCm, fn, imageName); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil

}

// Get Function Controller Configuration
func (r *ReconcileFunction) getFunctionControllerConfiguration(fnConfig *corev1.ConfigMap) error {

	if len(fnConfigNamespaceEnv) > 0 {
		fnConfigNamespace = fnConfigNamespaceEnv
	}

	if len(fnConfigNameEnv) > 0 {
		fnConfigName = fnConfigNameEnv
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: fnConfigName, Namespace: fnConfigNamespace}, fnConfig)
	if err != nil {
		log.Error(err, "Unable to read Function controller's configuration", "namespace", fnConfigNamespace, "name", fnConfigName)
		return err
	}

	// TODO REMOVE LOG
	log.Info("Function Controller's configuration found", "namespace", fnConfig.Namespace, "name", fnConfig.Name)

	return nil
}

// Get the Function instance
func (r *ReconcileFunction) getFunctionInstance(request reconcile.Request, fn *runtimev1alpha1.Function) error {
	// Get the Function instance
	err := r.Get(context.TODO(), request.NamespacedName, fn)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return nil
		}
		// Error reading the object - requeue the request.
		return err
	}

	// TODO REMOVE LOG
	log.Info("Function instance found:", "namespace", fn.Namespace, "name", fn.Name)

	return nil
}

func createFunctionHandlerMap(fn *runtimev1alpha1.Function) map[string]string {

	data := make(map[string]string)
	data["handler"] = "handler.main"
	data["handler.js"] = fn.Spec.Function
	if len(strings.Trim(fn.Spec.Deps, " ")) == 0 {
		data["package.json"] = "{}"
	} else {
		data["package.json"] = fn.Spec.Deps
	}

	return data

}

// Create Function's ConfigMap
func (r *ReconcileFunction) createFunctionConfigMap(foundCm *corev1.ConfigMap, deployCm *corev1.ConfigMap, fn *runtimev1alpha1.Function) error {

	// Create Function Handler
	deployCm.Data = createFunctionHandlerMap(fn)

	// Managing a ConfigMap
	deployCm.ObjectMeta = metav1.ObjectMeta{
		Labels:    fn.Labels,
		Namespace: fn.Namespace,
		Name:      fn.Name,
	}

	if err := controllerutil.SetControllerReference(fn, deployCm, r.scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: deployCm.Name, Namespace: deployCm.Namespace}, foundCm)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating the Function's ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		err = r.Create(context.TODO(), deployCm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO REMOVE LOG
	log.Info("Function ConfigMap:", "namespace", deployCm.Namespace, "name", deployCm.Name)

	return nil
}

// Update found Function's ConfigMap
func (r *ReconcileFunction) updateFunctionConfigMap(foundCm *corev1.ConfigMap, deployCm *corev1.ConfigMap) error {

	if !reflect.DeepEqual(deployCm.Data, foundCm.Data) {
		foundCm.Data = deployCm.Data
		log.Info("Updating Function's ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		err := r.Update(context.TODO(), foundCm)
		if err != nil {
			return err
		}
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: deployCm.Name, Namespace: deployCm.Namespace}, foundCm)
	if err != nil {
		log.Error(err, "Unable to read the updated Function ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)
		return err
	}

	// TODO REMOVE LOG
	log.Info("Updated Function'S ConfigMap", "namespace", deployCm.Namespace, "name", deployCm.Name)

	return nil

}

func (r *ReconcileFunction) getFunctionBuildTemplate(rnInfo *runtimeUtil.RuntimeInfo, fnConfig *corev1.ConfigMap, fn *runtimev1alpha1.Function, imageName string) error {

	if len(buildTemplateNameEnv) > 0 {
		buildTemplateName = buildTemplateNameEnv
	}

	if len(buildTemplateNamespaceEnv) > 0 {
		buildTemplateNamespace = buildTemplateNamespaceEnv
	}

	deployBuildTemplate := &buildv1alpha1.BuildTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.knative.dev/v1alpha1",
			Kind:       "Build",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildTemplateName,
			Namespace: buildTemplateNamespace,
		},
		Spec: runtimeUtil.GetBuildTemplateSpec(fn, imageName),
	}

	if err := controllerutil.SetControllerReference(fn, deployBuildTemplate, r.scheme); err != nil {
		return err
	}

	// Check if the BuildTemplate object already exists, if not create a new one.
	foundBuildTemplate := &buildv1alpha1.BuildTemplate{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deployBuildTemplate.Name, Namespace: deployBuildTemplate.Namespace}, foundBuildTemplate)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Knative BuildTemplate", "namespace", deployBuildTemplate.Namespace, "name", deployBuildTemplate.Name)
		err = r.Create(context.TODO(), deployBuildTemplate)
		if err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "Error while try to create Knative BuildTemplate", "namespace", deployBuildTemplate.Namespace, "name", deployBuildTemplate.Name)
		return err
	}

	if !reflect.DeepEqual(deployBuildTemplate.Spec, foundBuildTemplate.Spec) {
		foundBuildTemplate.Spec = deployBuildTemplate.Spec
		log.Info("Updating Knative BuildTemplate", "namespace", deployBuildTemplate.Namespace, "name", deployBuildTemplate.Name)
		err = r.Update(context.TODO(), foundBuildTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileFunction) buildFunctionImage(rnInfo *runtimeUtil.RuntimeInfo, fnConfig *corev1.ConfigMap, fn *runtimev1alpha1.Function, imageName string) error {

	// Create a new Build data structure
	build := runtimeUtil.NewBuild(rnInfo, fn, imageName)
	deployBuild := runtimeUtil.GetBuildResource(build, fn)
	if err := controllerutil.SetControllerReference(fn, deployBuild, r.scheme); err != nil {
		return err
	}

	// Check if the build object (building the function) already exists, if not create a new one.
	foundBuild := &buildv1alpha1.Build{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deployBuild.Name, Namespace: deployBuild.Namespace}, foundBuild)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		err = r.Create(context.TODO(), deployBuild)
		if err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "Error while try to create Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		return err
	}

	if !reflect.DeepEqual(deployBuild.Spec, foundBuild.Spec) {
		foundBuild.Spec = deployBuild.Spec
		log.Info("Updating Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		err = r.Update(context.TODO(), foundBuild)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileFunction) serveFunction(rnInfo *runtimeUtil.RuntimeInfo, foundCm *corev1.ConfigMap, fn *runtimev1alpha1.Function, imageName string) error {

	deployService := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    fn.Labels,
			Namespace: fn.Namespace,
			Name:      fn.Name,
		},
		Spec: runtimeUtil.GetServiceSpec(imageName, *fn, rnInfo),
	}

	if err := controllerutil.SetControllerReference(fn, deployService, r.scheme); err != nil {
		return err
	}

	// Check if the Serving object (serving the function) already exists, if not create a new one.
	foundService := &servingv1alpha1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deployService.Name, Namespace: deployService.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Knative Service", "namespace", deployService.Namespace, "name", deployService.Name)
		err = r.Create(context.TODO(), deployService)
		if err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "Error while try to create Knative Service", "namespace", deployService.Namespace, "name", deployService.Name)
		return err
	}

	if !reflect.DeepEqual(deployService.Spec, foundService.Spec) {
		foundService.Spec = deployService.Spec
		log.Info("Updating Knative Service", "namespace", deployService.Namespace, "name", deployService.Name)
		err = r.Update(context.TODO(), foundService)
		if err != nil {
			return err
		}
	}

	return nil
}
