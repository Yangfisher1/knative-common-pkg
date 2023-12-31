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

package v1

import (
	"testing"

	"github.com/Yangfisher1/knative-common-pkg/apis"
)

func TestGetConditionSet(t *testing.T) {
	testCases := []struct {
		name                      string
		condSet                   apis.ConditionSet
		expectedTopLevelCondition apis.ConditionType
	}{
		{
			name:                      "living set",
			condSet:                   apis.NewLivingConditionSet(),
			expectedTopLevelCondition: apis.ConditionReady,
		},
		{
			name:                      "batch set",
			condSet:                   apis.NewBatchConditionSet(),
			expectedTopLevelCondition: apis.ConditionSucceeded,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			resource := KResource{}
			test.condSet.Manage(resource.GetStatus()).InitializeConditions()

			mgr := resource.GetConditionSet().Manage(resource.GetStatus())

			if mgr.GetTopLevelCondition().Type != test.expectedTopLevelCondition {
				t.Errorf("wrong top-level condition got=%s want=%s",
					mgr.GetTopLevelCondition().Type, test.expectedTopLevelCondition)
			}
		})
	}
}
