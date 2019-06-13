package utils

import (
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	runtimev1alpha1 "github.com/kyma-incubator/runtime/pkg/apis/runtime/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

/*
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: sample
spec:
  template:
    metadata:
      annotations:
    spec:
      containers:
      - image: sample:latest
        env:
        - name:  "FUNC_HANDLER"
          value: "main"
        - name:  "MOD_NAME"
          value: "handler"
        - name:  "FUNC_TIMEOUT"
          value: "180"
        - name:  "FUNC_RUNTIME"
          value: "nodejs8"
        - name:  "FUNC_MEMORY_LIMIT"
          value: "128Mi"
        - name:  "FUNC_PORT"
          value: "8080"
        - name:  "NODE_PATH"
          value: "$(KUBELESS_INSTALL_VOLUME)/node_modules"
      requests: {} # to be filled in by the mutating admission controller
*/

// GetServiceSpec gets ServiceSpec for a function
func GetServiceSpec(imageName string, fn runtimev1alpha1.Function, rnInfo *RuntimeInfo) servingv1alpha1.ServiceSpec {

	// TODO: Make it constant for nodejs8/nodejs6
	envVarsForRevision := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}

	configuration := servingv1alpha1.ConfigurationSpec{
		Template: &servingv1alpha1.RevisionTemplateSpec{
			Spec: servingv1alpha1.RevisionSpec{
				RevisionSpec: v1beta1.RevisionSpec{
					PodSpec: v1beta1.PodSpec{
						Containers: []corev1.Container{{
							Image: imageName,
							Env:   envVarsForRevision,
						}},
					},
				},
			},
		},
	}

	return servingv1alpha1.ServiceSpec{
		ConfigurationSpec: configuration,
	}

}
