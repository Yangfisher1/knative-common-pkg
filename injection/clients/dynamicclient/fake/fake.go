/*
Copyright 2019 The Knative Authors

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

package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/Yangfisher1/knative-common-pkg/injection"
	"github.com/Yangfisher1/knative-common-pkg/injection/clients/dynamicclient"
	"github.com/Yangfisher1/knative-common-pkg/logging"
)

func init() {
	injection.Fake.RegisterClient(withClient)
}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	scheme := runtime.NewScheme()
	k8sscheme.AddToScheme(scheme)
	ctx, _ = With(ctx, scheme)
	return ctx
}

func With(ctx context.Context, scheme *runtime.Scheme, objects ...runtime.Object) (context.Context, *fake.FakeDynamicClient) {
	cs := fake.NewSimpleDynamicClient(scheme, objects...)
	return context.WithValue(ctx, dynamicclient.Key{}, cs), cs
}

// Get extracts the Kubernetes client from the context.
func Get(ctx context.Context) *fake.FakeDynamicClient {
	untyped := ctx.Value(dynamicclient.Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panicf(
			"Unable to fetch %T from context.", (*fake.FakeDynamicClient)(nil))
	}
	return untyped.(*fake.FakeDynamicClient)
}
