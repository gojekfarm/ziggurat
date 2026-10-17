package kafka

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

// Kafka security.protocol values relevant to SASL authentication. Set
// ConsumerConfig.SecurityProtocol to one of these explicitly, or leave it empty and
// set SASLMechanism (or just SASLUsername/SASLPassword) alone - toConfigMap then
// defaults to SecurityProtocolSASLPlaintext (see below).
//
// SASL_PLAINTEXT sends the SASL handshake (and therefore the SCRAM credentials, salted
// rather than in cleartext, but still over an unencrypted channel) without TLS. If your
// cluster terminates TLS and you want the connection itself encrypted too, set
// SecurityProtocol to SecurityProtocolSASLSSL explicitly (and set SSLCALocation if
// needed) - it is not the default here.
const (
	SecurityProtocolSASLPlaintext = "SASL_PLAINTEXT"
	SecurityProtocolSASLSSL       = "SASL_SSL"
)

// DefaultSASLMechanism is used by toConfigMap when SASLUsername/SASLPassword are set
// but SASLMechanism is left empty.
const DefaultSASLMechanism = SASLMechanismScramSHA512

// SASL mechanisms this package has been exercised against. librdkafka (the C library
// confluent-kafka-go binds to) implements SCRAM natively, so no extra system
// dependency (e.g. cyrus-sasl, needed for GSSAPI/Kerberos) is required to use these.
const (
	SASLMechanismScramSHA256 = "SCRAM-SHA-256"
	SASLMechanismScramSHA512 = "SCRAM-SHA-512"
)

type ConsumerConfig struct {
	BootstrapServers      string
	DebugLevel            string
	GroupID               string
	Topics                []string
	AutoCommitInterval    int
	ConsumerCount         int
	PollTimeout           int
	AutoOffsetReset       string
	PartitionAssignment   string
	MaxPollIntervalMS     int
	AllowAutoCreateTopics bool

	// SecurityProtocol, SASLMechanism, SASLUsername, SASLPassword, and SSLCALocation
	// configure SASL authentication (SCRAM-SHA-256/512) against a Kafka cluster that
	// enforces ACLs.
	//
	// Kafka ACLs are a broker-side concept: the broker authorizes operations against
	// the *authenticated principal* derived from the SASL handshake. This package has
	// no ACL logic of its own and can't - authorizing or denying an operation is the
	// broker's job. What this config does is let a ziggurat-go consumer authenticate
	// via SASL/SCRAM in the first place, so that whatever ACLs the cluster operator
	// has already configured for that principal actually apply to it, instead of the
	// client failing to connect at all (or connecting anonymously/unauthenticated,
	// if the broker allows that) and never reaching the ACL check.
	//
	// Leaving SASLMechanism empty while setting SASLUsername/SASLPassword defaults the
	// mechanism to DefaultSASLMechanism (SCRAM-SHA-512); leaving SecurityProtocol empty
	// defaults it to SecurityProtocolSASLPlaintext. Set either explicitly to override.
	//
	// SASLUsername/SASLPassword are credentials, not something to hardcode: read them
	// from an environment variable or your secrets manager of choice at startup, e.g.
	// os.Getenv("KAFKA_SASL_PASSWORD"). See the README for a full usage example.
	SecurityProtocol string
	SASLMechanism    string
	SASLUsername     string
	SASLPassword     string
	// SSLCALocation is the path to a CA certificate bundle used to verify the
	// broker's TLS certificate when SecurityProtocol is SecurityProtocolSASLSSL.
	// Leave empty to use librdkafka/OpenSSL's system default CA bundle, which is
	// correct for brokers with a certificate from a public CA; set it explicitly for
	// a private/internal CA.
	SSLCALocation string
}

func (c ConsumerConfig) toConfigMap() kafka.ConfigMap {

	kafkaConfMap := kafka.ConfigMap{
		"bootstrap.servers":        c.BootstrapServers,
		"group.id":                 c.GroupID,
		"enable.auto.commit":       true,
		"debug":                    "consumer",
		"go.logs.channel.enable":   true,
		"enable.auto.offset.store": false,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "latest",
	}

	if c.AutoOffsetReset != "" {
		kafkaConfMap["auto.offset.reset"] = c.AutoOffsetReset
	}

	if c.PartitionAssignment != "" {
		kafkaConfMap["partition.assignment.strategy"] = c.PartitionAssignment
	}

	if c.DebugLevel != "" {
		kafkaConfMap["debug"] = c.DebugLevel
	}

	if c.AutoCommitInterval > 0 {
		kafkaConfMap["auto.commit.interval.ms"] = c.AutoCommitInterval
	}

	if c.MaxPollIntervalMS > 0 {
		kafkaConfMap["max.poll.interval.ms"] = c.MaxPollIntervalMS
	}

	kafkaConfMap["allow.auto.create.topics"] = c.AllowAutoCreateTopics

	// SASL is considered "configured" if any of the three fields are set, not just
	// SASLMechanism - this lets a caller set only SASLUsername/SASLPassword and get
	// DefaultSASLMechanism rather than silently getting no SASL config at all.
	saslConfigured := c.SASLMechanism != "" || c.SASLUsername != "" || c.SASLPassword != ""

	if saslConfigured {
		mechanism := c.SASLMechanism
		if mechanism == "" {
			mechanism = DefaultSASLMechanism
		}
		kafkaConfMap["sasl.mechanisms"] = mechanism
		kafkaConfMap["sasl.username"] = c.SASLUsername
		kafkaConfMap["sasl.password"] = c.SASLPassword
	}

	switch {
	case c.SecurityProtocol != "":
		// Explicit setting always wins - e.g. SecurityProtocolSASLSSL for a cluster that
		// terminates TLS, without having to also touch the default below for every
		// environment that isn't using plain SASL_PLAINTEXT.
		kafkaConfMap["security.protocol"] = c.SecurityProtocol
	case saslConfigured:
		kafkaConfMap["security.protocol"] = SecurityProtocolSASLPlaintext
	}
	// If neither is set, security.protocol is left unset entirely, so librdkafka's own
	// default (PLAINTEXT) applies - zero config-map difference for existing callers
	// who don't use any of the fields above.

	if c.SSLCALocation != "" {
		kafkaConfMap["ssl.ca.location"] = c.SSLCALocation
	}

	return kafkaConfMap
}
