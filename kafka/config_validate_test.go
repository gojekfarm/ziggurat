package kafka

import "testing"

func TestValidate_NoSASLFields_Passes(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
	}.Validate()

	if err != nil {
		t.Errorf("expected no error when no SASL fields are set, got: %v", err)
	}
}

func TestValidate_FullValidSASLConfig_Passes(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA256,
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
		SecurityProtocol: SecurityProtocolSASLSSL,
	}.Validate()

	if err != nil {
		t.Errorf("expected no error for a fully-specified, consistent SASL config, got: %v", err)
	}
}

func TestValidate_UsernameOnly_Fails(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLUsername:     "alice",
	}.Validate()

	if err == nil {
		t.Fatal("expected an error when SASLUsername is set but SASLPassword is not")
	}
}

func TestValidate_PasswordOnly_Fails(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLPassword:     "s3cret",
	}.Validate()

	if err == nil {
		t.Fatal("expected an error when SASLPassword is set but SASLUsername is not")
	}
}

func TestValidate_MechanismOnlyNoCredentials_Fails(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    SASLMechanismScramSHA512,
	}.Validate()

	if err == nil {
		t.Fatal("expected an error when SASLMechanism is set but neither SASLUsername nor SASLPassword is")
	}
}

func TestValidate_UnsupportedSASLMechanism_Fails(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLMechanism:    "PLAIN",
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
	}.Validate()

	if err == nil {
		t.Fatal("expected an error for an unsupported SASLMechanism value")
	}
}

func TestValidate_SASLWithIncompatibleSecurityProtocol_Fails(t *testing.T) {
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SASLUsername:     "alice",
		SASLPassword:     "s3cret",
		SecurityProtocol: "PLAINTEXT",
	}.Validate()

	if err == nil {
		t.Fatal("expected an error when SASL is configured but SecurityProtocol is not a SASL-compatible value")
	}
}

func TestValidate_NonSASLSecurityProtocolAlone_Passes(t *testing.T) {
	// Plain TLS without SASL - SecurityProtocol is independent of SASL and shouldn't be
	// restricted to the SASL-compatible values when SASL isn't configured at all.
	err := ConsumerConfig{
		BootstrapServers: "localhost:9092",
		GroupID:          "g1",
		SecurityProtocol: "SSL",
	}.Validate()

	if err != nil {
		t.Errorf("expected no error for a non-SASL SecurityProtocol with no SASL fields set, got: %v", err)
	}
}
