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

package filtered

import (
	context "context"

	client "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/client"
	filtered "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/informers/factory/filtered"
	controller "github.com/Yangfisher1/knative-common-pkg/controller"
	injection "github.com/Yangfisher1/knative-common-pkg/injection"
	logging "github.com/Yangfisher1/knative-common-pkg/logging"
	apicertificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	v1beta1 "k8s.io/client-go/informers/certificates/v1beta1"
	kubernetes "k8s.io/client-go/kubernetes"
	certificatesv1beta1 "k8s.io/client-go/listers/certificates/v1beta1"
	cache "k8s.io/client-go/tools/cache"
)

func init() {
	injection.Default.RegisterFilteredInformers(withInformer)
	injection.Dynamic.RegisterDynamicInformer(withDynamicInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct {
	Selector string
}

func withInformer(ctx context.Context) (context.Context, []controller.Informer) {
	untyped := ctx.Value(filtered.LabelKey{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch labelkey from context.")
	}
	labelSelectors := untyped.([]string)
	infs := []controller.Informer{}
	for _, selector := range labelSelectors {
		f := filtered.Get(ctx, selector)
		inf := f.Certificates().V1beta1().CertificateSigningRequests()
		ctx = context.WithValue(ctx, Key{Selector: selector}, inf)
		infs = append(infs, inf.Informer())
	}
	return ctx, infs
}

func withDynamicInformer(ctx context.Context) context.Context {
	untyped := ctx.Value(filtered.LabelKey{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch labelkey from context.")
	}
	labelSelectors := untyped.([]string)
	for _, selector := range labelSelectors {
		inf := &wrapper{client: client.Get(ctx), selector: selector}
		ctx = context.WithValue(ctx, Key{Selector: selector}, inf)
	}
	return ctx
}

// Get extracts the typed informer from the context.
func Get(ctx context.Context, selector string) v1beta1.CertificateSigningRequestInformer {
	untyped := ctx.Value(Key{Selector: selector})
	if untyped == nil {
		logging.FromContext(ctx).Panicf(
			"Unable to fetch k8s.io/client-go/informers/certificates/v1beta1.CertificateSigningRequestInformer with selector %s from context.", selector)
	}
	return untyped.(v1beta1.CertificateSigningRequestInformer)
}

type wrapper struct {
	client kubernetes.Interface

	selector string
}

var _ v1beta1.CertificateSigningRequestInformer = (*wrapper)(nil)
var _ certificatesv1beta1.CertificateSigningRequestLister = (*wrapper)(nil)

func (w *wrapper) Informer() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(nil, &apicertificatesv1beta1.CertificateSigningRequest{}, 0, nil)
}

func (w *wrapper) Lister() certificatesv1beta1.CertificateSigningRequestLister {
	return w
}

func (w *wrapper) List(selector labels.Selector) (ret []*apicertificatesv1beta1.CertificateSigningRequest, err error) {
	reqs, err := labels.ParseToRequirements(w.selector)
	if err != nil {
		return nil, err
	}
	selector = selector.Add(reqs...)
	lo, err := w.client.CertificatesV1beta1().CertificateSigningRequests().List(context.TODO(), v1.ListOptions{
		LabelSelector: selector.String(),
		// TODO(mattmoor): Incorporate resourceVersion bounds based on staleness criteria.
	})
	if err != nil {
		return nil, err
	}
	for idx := range lo.Items {
		ret = append(ret, &lo.Items[idx])
	}
	return ret, nil
}

func (w *wrapper) Get(name string) (*apicertificatesv1beta1.CertificateSigningRequest, error) {
	// TODO(mattmoor): Check that the fetched object matches the selector.
	return w.client.CertificatesV1beta1().CertificateSigningRequests().Get(context.TODO(), name, v1.GetOptions{
		// TODO(mattmoor): Incorporate resourceVersion bounds based on staleness criteria.
	})
}
