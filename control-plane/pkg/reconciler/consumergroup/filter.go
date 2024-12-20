/*
 * Copyright 2021 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package consumergroup

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"

	kafkainternals "knative.dev/eventing-kafka-broker/control-plane/pkg/apis/internalskafkaeventing/v1alpha1"
)

// Filter returns a filter function based on the user-facing resource that a controller is tracking.
// Usable by FilteringResourceEventHandler.
func Filter(userFacingResource string) func(obj interface{}) bool {
	userFacingResource = strings.ToLower(userFacingResource)
	return func(obj interface{}) bool {
		cg, ok := obj.(*kafkainternals.ConsumerGroup)
		if !ok {
			return false
		}

		for _, or := range cg.OwnerReferences {
			if strings.ToLower(or.Kind) == userFacingResource {
				return true
			}
		}

		return false
	}
}

// Enqueue enqueues using the provided enqueue function the resource associated with a ConsumerGroup
func Enqueue(userFacingResource string, enqueue func(key types.NamespacedName)) func(obj interface{}) {
	userFacingResource = strings.ToLower(userFacingResource)
	return func(obj interface{}) {
		cg, ok := obj.(*kafkainternals.ConsumerGroup)
		if !ok {
			return
		}

		for _, or := range cg.OwnerReferences {
			if strings.ToLower(or.Kind) == userFacingResource {
				enqueue(types.NamespacedName{Namespace: cg.GetNamespace(), Name: or.Name})
			}
		}
	}
}
