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

package v1beta1

import (
	"context"
	"testing"

	"github.com/Yangfisher1/knative-common-pkg/apis"
	v1 "github.com/Yangfisher1/knative-common-pkg/apis/duck/v1"
	"github.com/google/go-cmp/cmp"
)

func TestConversion(t *testing.T) {
	tests := []struct {
		name        string
		addr        *Addressable
		conv        apis.Convertible
		want        string
		wantErrUp   bool
		wantErrDown bool
	}{{
		name: "v1",
		addr: &Addressable{
			URL: &apis.URL{
				Scheme: "https",
				Host:   "bar.com",
			},
		},
		conv:        &v1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name:        "v1 - empty",
		addr:        &Addressable{},
		conv:        &v1.Addressable{},
		wantErrUp:   false,
		wantErrDown: false,
	}, {
		name: "v1beta1",
		addr: &Addressable{
			URL: &apis.URL{
				Scheme: "https",
				Host:   "bar.com",
			},
		},
		conv:        &Addressable{},
		wantErrUp:   true,
		wantErrDown: true,
	}, {
		name:        "v1beta1 - empty",
		addr:        &Addressable{},
		conv:        &Addressable{},
		wantErrUp:   true,
		wantErrDown: true,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conv := test.conv
			if err := test.addr.ConvertTo(context.Background(), conv); err != nil {
				if !test.wantErrUp {
					t.Error("ConvertTo() =", err)
				}
			} else if test.wantErrUp {
				t.Errorf("ConvertTo() = %#v, wanted error", conv)
			}
			got := &Addressable{}
			if err := got.ConvertFrom(context.Background(), conv); err != nil {
				if !test.wantErrDown {
					t.Error("ConvertFrom() =", err)
				}
				return
			} else if test.wantErrDown {
				t.Errorf("ConvertFrom() = %#v, wanted error", conv)
				return
			}

			if diff := cmp.Diff(test.addr, got); diff != "" {
				t.Error("roundtrip (-want, +got) =", diff)
			}
		})
	}
}
