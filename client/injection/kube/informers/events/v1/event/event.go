/*
Copyright 2022 The Knative Authors

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

// Code generated by injection-gen. DO NOT EDIT.

package event

import (
	context "context"

	client "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/client"
	factory "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/informers/factory"
	controller "github.com/Yangfisher1/knative-common-pkg/controller"
	injection "github.com/Yangfisher1/knative-common-pkg/injection"
	logging "github.com/Yangfisher1/knative-common-pkg/logging"
	apieventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/informers/events/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	eventsv1 "k8s.io/client-go/listers/events/v1"
	cache "k8s.io/client-go/tools/cache"
)

func init() {
	injection.Default.RegisterInformer(withInformer)
	injection.Dynamic.RegisterDynamicInformer(withDynamicInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct{}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := factory.Get(ctx)
	inf := f.Events().V1().Events()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

func withDynamicInformer(ctx context.Context) context.Context {
	inf := &wrapper{client: client.Get(ctx), resourceVersion: injection.GetResourceVersion(ctx)}
	return context.WithValue(ctx, Key{}, inf)
}

// Get extracts the typed informer from the context.
func Get(ctx context.Context) v1.EventInformer {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch k8s.io/client-go/informers/events/v1.EventInformer from context.")
	}
	return untyped.(v1.EventInformer)
}

type wrapper struct {
	client kubernetes.Interface

	namespace string

	resourceVersion string
}

var _ v1.EventInformer = (*wrapper)(nil)
var _ eventsv1.EventLister = (*wrapper)(nil)

func (w *wrapper) Informer() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(nil, &apieventsv1.Event{}, 0, nil)
}

func (w *wrapper) Lister() eventsv1.EventLister {
	return w
}

func (w *wrapper) Events(namespace string) eventsv1.EventNamespaceLister {
	return &wrapper{client: w.client, namespace: namespace, resourceVersion: w.resourceVersion}
}

// SetResourceVersion allows consumers to adjust the minimum resourceVersion
// used by the underlying client.  It is not accessible via the standard
// lister interface, but can be accessed through a user-defined interface and
// an implementation check e.g. rvs, ok := foo.(ResourceVersionSetter)
func (w *wrapper) SetResourceVersion(resourceVersion string) {
	w.resourceVersion = resourceVersion
}

func (w *wrapper) List(selector labels.Selector) (ret []*apieventsv1.Event, err error) {
	lo, err := w.client.EventsV1().Events(w.namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector:   selector.String(),
		ResourceVersion: w.resourceVersion,
	})
	if err != nil {
		return nil, err
	}
	for idx := range lo.Items {
		ret = append(ret, &lo.Items[idx])
	}
	return ret, nil
}

func (w *wrapper) Get(name string) (*apieventsv1.Event, error) {
	return w.client.EventsV1().Events(w.namespace).Get(context.TODO(), name, metav1.GetOptions{
		ResourceVersion: w.resourceVersion,
	})
}
