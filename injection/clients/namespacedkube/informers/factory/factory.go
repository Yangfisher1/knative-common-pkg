/*
Copyright 2020 The Knative Authors

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

package factory

import (
	context "context"

	client "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/client"
	controller "github.com/Yangfisher1/knative-common-pkg/controller"
	injection "github.com/Yangfisher1/knative-common-pkg/injection"
	logging "github.com/Yangfisher1/knative-common-pkg/logging"
	"github.com/Yangfisher1/knative-common-pkg/system"
	informers "k8s.io/client-go/informers"
)

func init() {
	injection.Default.RegisterInformerFactory(withInformerFactory)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withInformerFactory(ctx context.Context) context.Context {
	c := client.Get(ctx)
	return context.WithValue(ctx, Key{},
		informers.NewSharedInformerFactoryWithOptions(c, controller.GetResyncPeriod(ctx),
			// This factory scopes things to the system namespace.
			informers.WithNamespace(system.Namespace())))
}

// Get extracts the InformerFactory from the context.
func Get(ctx context.Context) informers.SharedInformerFactory {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch k8s.io/client-go/informers.SharedInformerFactory from context.")
	}
	return untyped.(informers.SharedInformerFactory)
}
