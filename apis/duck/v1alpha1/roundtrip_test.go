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

package v1alpha1

import (
	"testing"

	pkgfuzzer "github.com/Yangfisher1/knative-common-pkg/apis/testing/fuzzer"
	"github.com/Yangfisher1/knative-common-pkg/apis/testing/roundtrip"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func TestEventingRoundTripTypesToJSON(t *testing.T) {
	scheme := runtime.NewScheme()
	utilruntime.Must(AddToScheme(scheme))

	fuzzerFuncs := fuzzer.MergeFuzzerFuncs(
		pkgfuzzer.Funcs,
	)
	roundtrip.ExternalTypesViaJSON(t, scheme, fuzzerFuncs)
}
