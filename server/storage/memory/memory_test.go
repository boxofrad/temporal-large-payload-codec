package memory_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/DataDog/temporal-large-payload-codec/server/storage"
	"github.com/DataDog/temporal-large-payload-codec/server/storage/memory"
)

func TestDriver_PutPayload(t *testing.T) {
	var err error
	var (
		ctx = context.Background()
		d   = memory.Driver{}
		buf = bytes.Buffer{}
	)

	// Get missing payload
	_, err = d.GetPayload(ctx, &storage.GetRequest{Digest: "sha256:foobar", Writer: &buf})
	var blobNotFound *storage.ErrBlobNotFound
	if !errors.As(err, &blobNotFound) {
		t.Errorf("expected error %q, got %q", storage.ErrBlobNotFound{}, err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no bytes to be written, got %d", buf.Len())
	}

	// Put a payload
	testPayloadBytes := []byte("hello world")
	if _, err := d.PutPayload(ctx, &storage.PutRequest{
		Data:          bytes.NewReader(testPayloadBytes),
		Digest:        "sha256:test",
		ContentLength: uint64(len(testPayloadBytes)),
	}); err != nil {
		t.Fatal(err)
	}

	// Get the payload back out and compare to original bytes
	_, err = d.GetPayload(ctx, &storage.GetRequest{Digest: "sha256:test", Writer: &buf})
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(testPayloadBytes) {
		t.Errorf("expected payload data %q, got %q", testPayloadBytes, b)
	}

	// Delete the payload 
	_, err = d.DeletePayload(ctx, &storage.DeleteRequest{
		Keys: []string{"sha256:test"},
	}) 
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.GetPayload(ctx, &storage.GetRequest{Digest: "sha256:test", Writer: &buf})
	if err == nil {
		t.Errorf("expected payload to be deleted")
	}
}
