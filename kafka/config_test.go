package kafka

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func TestToConfigMap_NoSASLFields_LeavesSecurityPropertiesUnset(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
	}.toConfigMap()

	for _, key := range []string{"security.protocol", "sasl.mechanisms", "sasl.username", "sasl.password", "ssl.ca.location"} {
		if _, ok := cm[key]; ok {
			t.Errorf("expected %q to be unset when no SASL fields are configured, got %v", key, cm[key])
		}
	}
}

func TestToConfigMap_SASLConfigured_DefaultsSecurityProtocolToSASLPlaintext(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA256,
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
	}.toConfigMap()

	assertConfigValue(t, cm, "security.protocol", SecurityProtocolSASLPlaintext)
	assertConfigValue(t, cm, "sasl.mechanisms", SASLMechanismScramSHA256)
	assertConfigValue(t, cm, "sasl.username", "alice")
	assertConfigValue(t, cm, "sasl.password", "s3cret")
}

func TestToConfigMap_SASLMechanism_DefaultsToScramSHA512(t *testing.T) {
	// Only username/password set, no explicit mechanism.
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
	}.toConfigMap()

	assertConfigValue(t, cm, "sasl.mechanisms", SASLMechanismScramSHA512)
	assertConfigValue(t, cm, "security.protocol", SecurityProtocolSASLPlaintext)
}

func TestToConfigMap_ExplicitSASLMechanism_OverridesDefault(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA256,
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
	}.toConfigMap()

	assertConfigValue(t, cm, "sasl.mechanisms", SASLMechanismScramSHA256)
}

func TestToConfigMap_ExplicitSecurityProtocol_OverridesDefault(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA512,
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
		// Default is SecurityProtocolSASLPlaintext - set SASL_SSL explicitly to get TLS too.
		SecurityProtocol: SecurityProtocolSASLSSL,
	}.toConfigMap()

	assertConfigValue(t, cm, "security.protocol", SecurityProtocolSASLSSL)
	assertConfigValue(t, cm, "sasl.mechanisms", SASLMechanismScramSHA512)
}

func TestToConfigMap_SecurityProtocolWithoutSASL_StillApplied(t *testing.T) {
	// e.g. plain TLS without SASL - SecurityProtocol is independent of SASLMechanism.
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SecurityProtocol: "SSL",
	}.toConfigMap()

	assertConfigValue(t, cm, "security.protocol", "SSL")
	if _, ok := cm["sasl.mechanisms"]; ok {
		t.Errorf("expected sasl.mechanisms to be unset when SASLMechanism isn't configured")
	}
}

func TestToConfigMap_SSLCALocation_PassedThrough(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA256,
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
		SSLCALocation:    "/etc/ssl/certs/ca.pem",
	}.toConfigMap()

	assertConfigValue(t, cm, "ssl.ca.location", "/etc/ssl/certs/ca.pem")
}

func TestToConfigMap_NoClientRack_LeavesClientRackUnset(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
	}.toConfigMap()

	if _, ok := cm["client.rack"]; ok {
		t.Errorf("expected client.rack to be unset when ClientRack isn't configured, got %v", cm["client.rack"])
	}
}

func TestToConfigMap_ClientRack_PassedThrough(t *testing.T) {
	cm := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		ClientRack:       "us-east-1a",
	}.toConfigMap()

	assertConfigValue(t, cm, "client.rack", "us-east-1a")
}

func assertConfigValue(t *testing.T, cm kafka.ConfigMap, key string, want interface{}) {
	t.Helper()
	got, ok := cm[key]
	if !ok {
		t.Fatalf("expected %q to be set, got nothing", key)
	}
	if got != want {
		t.Errorf("expected %q to be %v, got %v", key, want, got)
	}
}
