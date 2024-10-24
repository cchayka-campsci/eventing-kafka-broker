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
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	"knative.dev/eventing/pkg/apis/duck"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	// KafkaConditionReady has status True when the KafkaSource is ready to send events.
	KafkaConditionReady = apis.ConditionReady

	// KafkaConditionSinkProvided has status True when the KafkaSource has been configured with a sink target.
	KafkaConditionSinkProvided apis.ConditionType = "SinkProvided"

	// KafkaConditionDeployed has status True when the KafkaSource has had it's receive adapter deployment created.
	KafkaConditionDeployed apis.ConditionType = "Deployed"

	// KafkaConditionKeyType is True when the KafkaSource has been configured with valid key type for
	// the key deserializer.
	KafkaConditionKeyType apis.ConditionType = "KeyTypeCorrect"

	// KafkaConditionConnectionEstablished has status True when the Kafka configuration to use by the source
	// succeeded in establishing a connection to Kafka.
	KafkaConditionConnectionEstablished apis.ConditionType = "ConnectionEstablished"

	// KafkaConditionInitialOffsetsCommitted is True when the KafkaSource has committed the
	// initial offset of all claims
	KafkaConditionInitialOffsetsCommitted apis.ConditionType = "InitialOffsetsCommitted"

	// KafkaConditionOIDCIdentityCreated has status True when the KafkaSource has created an OIDC identity.
	KafkaConditionOIDCIdentityCreated apis.ConditionType = "OIDCIdentityCreated"
)

var (
	KafkaSourceCondSet = apis.NewLivingConditionSet(
		KafkaConditionSinkProvided,
		KafkaConditionDeployed,
		KafkaConditionConnectionEstablished,
		KafkaConditionInitialOffsetsCommitted,
		KafkaConditionOIDCIdentityCreated,
	)

	kafkaCondSetLock = sync.RWMutex{}
)

// RegisterAlternateKafkaConditionSet register an alternate apis.ConditionSet.
func RegisterAlternateKafkaConditionSet(conditionSet apis.ConditionSet) {
	kafkaCondSetLock.Lock()
	defer kafkaCondSetLock.Unlock()

	KafkaSourceCondSet = conditionSet
}

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*KafkaSource) GetConditionSet() apis.ConditionSet {
	return KafkaSourceCondSet
}

func (s *KafkaSourceStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return KafkaSourceCondSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *KafkaSourceStatus) IsReady() bool {
	return KafkaSourceCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *KafkaSourceStatus) InitializeConditions() {
	KafkaSourceCondSet.Manage(s).InitializeConditions()
}

// MarkSink sets the condition that the source has a sink configured.
func (s *KafkaSourceStatus) MarkSink(addr *duckv1.Addressable) {
	if addr.URL != nil && !addr.URL.IsEmpty() {
		s.SinkURI = addr.URL
		s.SinkCACerts = addr.CACerts
		s.SinkAudience = addr.Audience
		KafkaSourceCondSet.Manage(s).MarkTrue(KafkaConditionSinkProvided)
	} else {
		KafkaSourceCondSet.Manage(s).MarkUnknown(KafkaConditionSinkProvided, "SinkEmpty", "Sink has resolved to empty.%s", "")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (s *KafkaSourceStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionSinkProvided, reason, messageFormat, messageA...)
}

func DeploymentIsAvailable(d *appsv1.DeploymentStatus, def bool) bool {
	// Check if the Deployment is available.
	for _, cond := range d.Conditions {
		if cond.Type == appsv1.DeploymentAvailable {
			return cond.Status == "True"
		}
	}
	return def
}

// MarkDeployed sets the condition that the source has been deployed.
func (s *KafkaSourceStatus) MarkDeployed(d *appsv1.Deployment) {
	if duck.DeploymentIsAvailable(&d.Status, false) {
		KafkaSourceCondSet.Manage(s).MarkTrue(KafkaConditionDeployed)

		// Propagate the number of consumers
		s.Consumers = d.Status.Replicas
	} else {
		// I don't know how to propagate the status well, so just give the name of the Deployment
		// for now.
		KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionDeployed, "DeploymentUnavailable", "The Deployment '%s' is unavailable.", d.Name)
	}
}

// MarkDeploying sets the condition that the source is deploying.
func (s *KafkaSourceStatus) MarkDeploying(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkUnknown(KafkaConditionDeployed, reason, messageFormat, messageA...)
}

// MarkNotDeployed sets the condition that the source has not been deployed.
func (s *KafkaSourceStatus) MarkNotDeployed(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionDeployed, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) MarkKeyTypeCorrect() {
	KafkaSourceCondSet.Manage(s).MarkTrue(KafkaConditionKeyType)
}

func (s *KafkaSourceStatus) MarkKeyTypeIncorrect(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionKeyType, reason, messageFormat, messageA...)
}

func (cs *KafkaSourceStatus) MarkConnectionEstablished() {
	KafkaSourceCondSet.Manage(cs).MarkTrue(KafkaConditionConnectionEstablished)
}

func (cs *KafkaSourceStatus) MarkConnectionNotEstablished(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(cs).MarkFalse(KafkaConditionConnectionEstablished, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) MarkInitialOffsetCommitted() {
	KafkaSourceCondSet.Manage(s).MarkTrue(KafkaConditionInitialOffsetsCommitted)
}

func (s *KafkaSourceStatus) MarkInitialOffsetNotCommitted(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionInitialOffsetsCommitted, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) MarkOIDCIdentityCreatedSucceeded() {
	KafkaSourceCondSet.Manage(s).MarkTrue(KafkaConditionOIDCIdentityCreated)
}

func (s *KafkaSourceStatus) MarkOIDCIdentityCreatedSucceededWithReason(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkTrueWithReason(KafkaConditionOIDCIdentityCreated, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) MarkOIDCIdentityCreatedFailed(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkFalse(KafkaConditionOIDCIdentityCreated, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) MarkOIDCIdentityCreatedUnknown(reason, messageFormat string, messageA ...interface{}) {
	KafkaSourceCondSet.Manage(s).MarkUnknown(KafkaConditionOIDCIdentityCreated, reason, messageFormat, messageA...)
}

func (s *KafkaSourceStatus) UpdateConsumerGroupStatus(status string) {
	s.Claims = status
}
