package rabbitmq

import "testing"

func TestWithPublisherConfirms_DefaultsBufferWhenZero(t *testing.T) {
	r := AutoRetry(nil, WithPublisherConfirms(0))

	if !r.publisherConfirms {
		t.Fatalf("expected publisher confirms to be enabled")
	}
	if r.publisherConfirmsBuffer != defaultPublisherConfirmsBuffer {
		t.Fatalf("expected default buffer %d, got %d", defaultPublisherConfirmsBuffer, r.publisherConfirmsBuffer)
	}
}

func TestWithPublisherConfirms_HonorsExplicitBuffer(t *testing.T) {
	r := AutoRetry(nil, WithPublisherConfirms(42))

	if !r.publisherConfirms {
		t.Fatalf("expected publisher confirms to be enabled")
	}
	if r.publisherConfirmsBuffer != 42 {
		t.Fatalf("expected buffer 42, got %d", r.publisherConfirmsBuffer)
	}
}

func TestARetry_publisherOptions_EmptyWhenDisabled(t *testing.T) {
	r := AutoRetry(nil)

	if opts := r.publisherOptions(); len(opts) != 0 {
		t.Fatalf("expected no publisher options when confirms are disabled, got %d", len(opts))
	}
}

func TestARetry_publisherOptions_SetWhenEnabled(t *testing.T) {
	r := AutoRetry(nil, WithPublisherConfirms(10))

	opts := r.publisherOptions()
	if len(opts) != 1 {
		t.Fatalf("expected exactly one publisher option when confirms are enabled, got %d", len(opts))
	}
}
