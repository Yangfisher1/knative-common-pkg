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
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
)

func TestStatusGetCondition(t *testing.T) {
	foo := &apis.Condition{
		Type:    "Foo",
		Status:  corev1.ConditionTrue,
		Message: "Something something foo",
	}
	bar := &apis.Condition{
		Type:    "Bar",
		Status:  corev1.ConditionTrue,
		Message: "Something something bar",
	}
	s := &Status{
		Conditions: Conditions{*foo, *bar},
	}

	got := s.GetCondition(foo.Type)
	if diff := cmp.Diff(got, foo); diff != "" {
		t.Error("GetCondition(foo) =", diff)
	}

	got = s.GetCondition(bar.Type)
	if diff := cmp.Diff(got, bar); diff != "" {
		t.Error("GetCondition(bar) =", diff)
	}

	if got := s.GetCondition("None"); got != nil {
		t.Errorf("GetCondition(None) = %v, wanted nil", got)
	}
}

func TestConditionSet(t *testing.T) {
	condSet := apis.NewLivingConditionSet("Foo")

	const wantGeneration = 42

	s := &Status{
		ObservedGeneration: wantGeneration,
		Annotations: map[string]string{
			"burning": "the",
			"bridges": "down",
		},
	}
	mgr := condSet.Manage(s)

	mgr.InitializeConditions()

	for _, c := range []apis.ConditionType{apis.ConditionReady, "Foo"} {
		if cond := mgr.GetCondition(c); cond == nil {
			t.Errorf("GetCondition(%q) = nil, wanted non-nil", c)
		} else if got, want := cond.Status, corev1.ConditionUnknown; got != want {
			t.Errorf("GetCondition(%q) = %v, wanted %v", c, got, want)
		}
	}

	s2 := &Status{}
	s.ConvertTo(context.Background(), s2)
	if condSet.Manage(s2).IsHappy() {
		t.Error("s2.IsHappy() = true, wanted false")
	}
	if got, want := len(s2.Conditions), 1; got != want {
		t.Errorf("len(s2.Conditions) = %d, wanted %d", got, want)
	}
	if gotGeneration := s2.ObservedGeneration; wantGeneration != gotGeneration {
		t.Errorf("s2.ObservedGeneration = %d, wanted %d",
			gotGeneration, wantGeneration)
	}
	if got, want := s2.Annotations, s.Annotations; !cmp.Equal(got, want) {
		t.Errorf("Annotations mismatch: diff(-want,+got):\n%s", cmp.Diff(want, got))
	}

	for _, c := range []apis.ConditionType{"Foo"} {
		mgr.MarkFalse(c, "bad", "for business")
	}

	for _, c := range []apis.ConditionType{apis.ConditionReady, "Foo"} {
		if cond := mgr.GetCondition(c); cond == nil {
			t.Errorf("GetCondition(%q) = nil, wanted non-nil", c)
		} else if got, want := cond.Status, corev1.ConditionFalse; got != want {
			t.Errorf("GetCondition(%q) = %v, wanted %v", c, got, want)
		}
	}

	s2 = &Status{}
	s.ConvertTo(context.Background(), s2)
	if condSet.Manage(s2).IsHappy() {
		t.Error("s2.IsHappy() = true, wanted false")
	}
	if got, want := len(s2.Conditions), 1; got != want {
		t.Errorf("len(s2.Conditions) = %d, wanted %d", got, want)
	}
	if gotGeneration := s2.ObservedGeneration; wantGeneration != gotGeneration {
		t.Errorf("len(s2.ObservedGeneration) = %d, wanted %d",
			gotGeneration, wantGeneration)
	}

	for _, c := range []apis.ConditionType{"Foo"} {
		mgr.MarkTrue(c)
	}

	for _, c := range []apis.ConditionType{apis.ConditionReady, "Foo"} {
		if cond := mgr.GetCondition(c); cond == nil {
			t.Errorf("GetCondition(%q) = nil, wanted non-nil", c)
		} else if got, want := cond.Status, corev1.ConditionTrue; got != want {
			t.Errorf("GetCondition(%q) = %v, wanted %v", c, got, want)
		}
	}

	s2 = &Status{}
	s.ConvertTo(context.Background(), s2)
	if !condSet.Manage(s2).IsHappy() {
		t.Error("s2.IsHappy() = false, wanted true")
	}
	if got, want := len(s2.Conditions), 1; got != want {
		t.Errorf("len(s2.Conditions) = %d, wanted %d", got, want)
	}
	if gotGeneration := s2.ObservedGeneration; wantGeneration != gotGeneration {
		t.Errorf("len(s2.ObservedGeneration) = %d, wanted %d",
			gotGeneration, wantGeneration)
	}
	s.Annotations = nil
	s2 = &Status{}
	s.ConvertTo(context.Background(), s2)
	if s2.Annotations != nil {
		t.Error("Annotations were not nil:", s2.Annotations)
	}
}
