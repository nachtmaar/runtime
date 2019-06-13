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
	"os"
	"testing"
	"time"
	runtimev1alpha1 "github.com/kyma-incubator/runtime/pkg/apis/runtime/v1alpha1"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var depKey = types.NamespacedName{Name: "foo", Namespace: "default"}

const timeout = time.Second * 10

func TestReconcile(t *testing.T) {
	//g := gomega.NewGomegaWithT(t)
	//fnCreated := &runtimev1alpha1.Function{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "foo",
	//		Namespace: "default",
	//	},
	//	Spec: runtimev1alpha1.FunctionSpec{
	//		Function:            "main() {asdfasdf}",
	//		FunctionContentType: "plaintext",
	//		Size:                "L",
	//		Runtime:             "nodejs6",
	//	},
	//}
	//
	//expectedEnv := []corev1.EnvVar{
	//	{
	//		Name:  "FUNC_HANDLER",
	//		Value: "main",
	//	},
	//	{
	//		Name:  "MOD_NAME",
	//		Value: "handler",
	//	},
	//	{
	//		Name:  "FUNC_TIMEOUT",
	//		Value: "180",
	//	},
	//	{
	//		Name:  "FUNC_RUNTIME",
	//		Value: "nodejs8",
	//	},
	//	{
	//		Name:  "FUNC_MEMORY_LIMIT",
	//		Value: "128Mi",
	//	},
	//	{
	//		Name:  "FUNC_PORT",
	//		Value: "8080",
	//	},
	//	{
	//		Name:  "NODE_PATH",
	//		Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
	//	},
	//}
	//
	//fnConfig := &corev1.ConfigMap{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "fn-config",
	//		Namespace: "default",
	//	},
	//	Data: map[string]string{
	//		"dockerRegistry":     "test",
	//		"serviceAccountName": "build-bot",
	//		"runtimes": `[
	//			{
	//				"ID": "nodejs8",
	//				"DockerFileName": "dockerfile-nodejs8",
	//			}
	//		]`,
	//	},
	//}
	//
	//dockerFileConfigNodejs := &corev1.ConfigMap{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "dockerfile-nodejs8",
	//		Namespace: "default",
	//	},
	//	Data: map[string]string{
	//		"Dockerfile": `FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
	//			USER root
	//			RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
	//				mkdir /kubeless && \
	//				cp /src/handler.js /kubeless && \
	//				cp /src/package.json /kubeless && \
	//				/kubeless-npm-install.sh
	//			USER 1000
	//		`,
	//	},
	//}
	//mgr, err := manager.New(cfg, manager.Options{})
	//g.Expect(err).NotTo(gomega.HaveOccurred())
	//c = mgr.GetClient()
	//
	//recFn, requests := SetupTestReconcile(newReconciler(mgr))
	//g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	//
	//stopMgr, mgrStopped := StartTestManager(mgr, g)
	//
	//defer func() {
	//	close(stopMgr)
	//	mgrStopped.Wait()
	//}()
	//
	//// Create the Function object and expect the Reconcile and Deployment to be created
	//err = c.Create(context.TODO(), dockerFileConfigNodejs)
	//if apierrors.IsInvalid(err) {
	//	t.Logf("failed to create object, got an invalid object error: %v", err)
	//	return
	//}
	//
	//err = c.Create(context.TODO(), fnConfig)
	//if apierrors.IsInvalid(err) {
	//	t.Logf("failed to create object, got an invalid object error: %v", err)
	//	return
	//}
	//
	//err = c.Create(context.TODO(), fnCreated)
	//if apierrors.IsInvalid(err) {
	//	t.Logf("failed to create object, got an invalid object error: %v", err)
	//	return
	//}
	//g.Expect(err).NotTo(gomega.HaveOccurred())
	//
	//g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//
	//cm := &corev1.ConfigMap{
	//	ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
	//}
	//
	//service := &servingv1alpha1.Service{
	//	ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
	//}
	//
	//g.Eventually(func() error { return c.Get(context.TODO(), depKey, cm) }, timeout).
	//	Should(gomega.Succeed())
	//
	//g.Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).
	//	Should(gomega.Succeed())
	//g.Expect(service.Namespace).To(gomega.Equal("default"))
	//
	//g.Expect(service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env).To(gomega.Equal(expectedEnv))
	//build := (*service.Spec.RunLatest.Configuration.Build)
	//buildByte, err := build.MarshalJSON()
	//if err != nil {
	//	t.Fatalf("Error while marshaling build object: %v", err)
	//}
	//var buildSpec buildv1alpha1.BuildSpec
	//err = json.Unmarshal(buildByte, &buildSpec)
	//if err != nil {
	//	t.Fatalf("Error while unmarshaling buildSpec: %v", err)
	//}
	//g.Expect(service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image).To(gomega.HavePrefix("test/default-foo"))
	//g.Expect(len(buildSpec.Volumes)).To(gomega.Equal(2))
	//g.Expect(buildSpec.ServiceAccountName).To(gomega.Equal("build-bot"))
	//g.Expect(service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image).To(gomega.HavePrefix("test/default-foo"))
	//
	//fnFetched := &runtimev1alpha1.Function{}
	//g.Expect(c.Get(context.TODO(), depKey, fnFetched)).NotTo(gomega.HaveOccurred())
	//g.Expect(fnFetched.Spec).To(gomega.Equal(fnCreated.Spec))
	//
	//fnUpdated := fnFetched.DeepCopy()
	//fnUpdated.Spec.Function = `main() {return "bla"}`
	//fnUpdated.Spec.Deps = `dependencies`
	//
	//fnFetched = &runtimev1alpha1.Function{}
	//g.Expect(c.Update(context.TODO(), fnUpdated)).NotTo(gomega.HaveOccurred())
	//g.Expect(c.Get(context.TODO(), depKey, fnFetched)).NotTo(gomega.HaveOccurred())
	//g.Expect(fnFetched.Spec).To(gomega.Equal(fnUpdated.Spec))
	//
	//g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//
	//cmUpdated := &corev1.ConfigMap{}
	//g.Eventually(func() string {
	//	c.Get(context.TODO(), depKey, cmUpdated)
	//	return cmUpdated.Data["handler.js"]
	//}, timeout, 1*time.Second).Should(gomega.Equal(fnUpdated.Spec.Function))
	//g.Eventually(func() string {
	//	c.Get(context.TODO(), depKey, cmUpdated)
	//	return cmUpdated.Data["package.json"]
	//}, timeout, 1*time.Second).Should(gomega.Equal(`dependencies`))
	//
	//ksvcUpdated := &servingv1alpha1.Service{}
	//g.Expect(c.Get(context.TODO(), depKey, ksvcUpdated)).NotTo(gomega.HaveOccurred())
	//
	//hash := sha256.New()
	//hash.Write([]byte(cmUpdated.Data["handler.js"] + cmUpdated.Data["package.json"]))
	//functionSha := fmt.Sprintf("%x", hash.Sum(nil))
	//
	//g.Expect(ksvcUpdated.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image).
	//	To(gomega.Equal(fmt.Sprintf("test/%s-%s:%s", "default", "foo", functionSha)))
}

func TestReconcileErrors(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	fnCreated := &runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "errortest",
			Namespace: "default",
		},
		Spec: runtimev1alpha1.FunctionSpec{
			Function:            "main() {asdfasdf}",
			FunctionContentType: "plaintext",
			Size:                "L",
			Runtime:             "nodejs6",
		},
	}

	fnConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fn-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"dockerRegistry":     "test",
			"serviceAccountName": "build-bot",
			"runtimes": `[
				{
					"ID": "nodejs8",
					"DockerFileName": "dockerfile-nodejs8",
				}
			]`,
		},
	}

	dockerFileConfigNodejs := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dockerfile-nodejs8",
			Namespace: "default",
		},
		Data: map[string]string{
			"Dockerfile": `FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
				USER root
				RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
					mkdir /kubeless && \
					cp /src/handler.js /kubeless && \
					cp /src/package.json /kubeless && \
					/kubeless-npm-install.sh
				USER 1000
			`,
		},
	}

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	// Create the Function object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), dockerFileConfigNodejs)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}

	err = c.Create(context.TODO(), fnConfig)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}

	err = c.Create(context.TODO(), fnCreated)
	t.Logf("%+v", err)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "bla",
			Namespace: "blabla",
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(runtimev1alpha1.SchemeGroupVersion, fnCreated)
	r := &ReconcileFunction{Client: c, scheme: s}

	os.Setenv("CONTROLLER_CONFIGMAP", "bla-config")
	_, err = r.Reconcile(request)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal(`ConfigMap "bla-config" not found`))

	os.Unsetenv("CONTROLLER_CONFIGMAP")
	os.Setenv("CONTROLLER_CONFIGMAP_NS", "stage")
	_, err = r.Reconcile(request)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal(`ConfigMap "fn-config" not found`))

}
