package core

import (
	"context"
	"sort"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

// StreamInfo contains information about a JetStream stream.
type StreamInfo struct {
	Name       string
	Messages   uint64
	Bytes      uint64
	Consumers  int
	Created    string
	FirstSeq   uint64
	LastSeq    uint64
	NumSubjects uint64
}

// KVBucketInfo contains information about a JetStream KV bucket.
type KVBucketInfo struct {
	Name    string
	Keys    uint64
	Bytes   uint64
	History int64
	TTL     string
}

// ObjectStoreInfo contains information about a JetStream object store.
type ObjectStoreInfo struct {
	Name   string
	Size   uint64
	Sealed bool
}

// SystemInfo contains overall system information.
type SystemInfo struct {
	Streams      []StreamInfo
	KVBuckets    []KVBucketInfo
	ObjectStores []ObjectStoreInfo
}

// GetSystemInfo retrieves system information about NATS/JetStream.
func (c *ChattoCore) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info := &SystemInfo{
		Streams:      []StreamInfo{},
		KVBuckets:    []KVBucketInfo{},
		ObjectStores: []ObjectStoreInfo{},
	}

	// List all streams, excluding KV and Object Store backing streams
	// (those are shown separately in their respective tables)
	streamLister := c.js.ListStreams(ctx)
	for si := range streamLister.Info() {
		// Skip KV backing streams (prefix "KV_") and Object Store backing streams (prefix "OBJ_")
		if strings.HasPrefix(si.Config.Name, "KV_") || strings.HasPrefix(si.Config.Name, "OBJ_") {
			continue
		}
		info.Streams = append(info.Streams, StreamInfo{
			Name:        si.Config.Name,
			Messages:    si.State.Msgs,
			Bytes:       si.State.Bytes,
			Consumers:   si.State.Consumers,
			Created:     si.Created.Format("2006-01-02 15:04:05"),
			FirstSeq:    si.State.FirstSeq,
			LastSeq:     si.State.LastSeq,
			NumSubjects: si.State.NumSubjects,
		})
	}

	// List all KV buckets
	kvLister := c.js.KeyValueStores(ctx)
	for status := range kvLister.Status() {
		ttlStr := "none"
		if status.TTL() > 0 {
			ttlStr = status.TTL().String()
		}
		info.KVBuckets = append(info.KVBuckets, KVBucketInfo{
			Name:    status.Bucket(),
			Keys:    status.Values(),
			Bytes:   status.Bytes(),
			History: status.History(),
			TTL:     ttlStr,
		})
	}

	// List all object stores
	osLister := c.js.ObjectStores(ctx)
	for status := range osLister.Status() {
		info.ObjectStores = append(info.ObjectStores, ObjectStoreInfo{
			Name:   status.Bucket(),
			Size:   status.Size(),
			Sealed: status.Sealed(),
		})
	}

	return info, nil
}

// GetConnectionInfo returns information about the NATS connection.
type ConnectionInfo struct {
	Connected    bool
	ServerID     string
	ServerName   string
	Version      string
	MaxPayload   int64
	RTT          string
}

// GetConnectionInfo retrieves NATS connection information.
func (c *ChattoCore) GetConnectionInfo() *ConnectionInfo {
	info := &ConnectionInfo{
		Connected: c.nc.IsConnected(),
	}

	if c.nc.IsConnected() {
		info.ServerID = c.nc.ConnectedServerId()
		info.ServerName = c.nc.ConnectedServerName()
		info.Version = c.nc.ConnectedServerVersion()
		info.MaxPayload = c.nc.MaxPayload()

		if rtt, err := c.nc.RTT(); err == nil {
			info.RTT = rtt.String()
		}
	}

	return info
}

// AccountInfo contains JetStream account limits and usage.
type AccountInfo struct {
	Memory         uint64
	MemoryUsed     uint64
	Storage        uint64
	StorageUsed    uint64
	Streams        int
	StreamsUsed    int
	Consumers      int
	ConsumersUsed  int
}

// GetAccountInfo retrieves JetStream account information.
func (c *ChattoCore) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	acc, err := c.js.AccountInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &AccountInfo{
		Memory:        uint64(acc.Limits.MaxMemory),
		MemoryUsed:    acc.Memory,
		Storage:       uint64(acc.Limits.MaxStore),
		StorageUsed:   acc.Store,
		Streams:       acc.Limits.MaxStreams,
		StreamsUsed:   acc.Streams,
		Consumers:     acc.Limits.MaxConsumers,
		ConsumersUsed: acc.Consumers,
	}, nil
}

// StreamSubject represents a subject in a stream with its message count.
type StreamSubject struct {
	Subject  string
	Messages uint64
}

// GetStreamSubjects retrieves all subjects and their message counts for a stream.
func (c *ChattoCore) GetStreamSubjects(ctx context.Context, streamName string) ([]StreamSubject, error) {
	stream, err := c.js.Stream(ctx, streamName)
	if err != nil {
		return nil, err
	}

	// Use WithSubjectFilter(">") to get all subjects with their message counts
	info, err := stream.Info(ctx, jetstream.WithSubjectFilter(">"))
	if err != nil {
		return nil, err
	}

	subjects := make([]StreamSubject, 0, len(info.State.Subjects))
	for subject, count := range info.State.Subjects {
		subjects = append(subjects, StreamSubject{
			Subject:  subject,
			Messages: count,
		})
	}

	// Sort by subject name for consistent ordering
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].Subject < subjects[j].Subject
	})

	return subjects, nil
}

// GetKVKeys retrieves all keys for a KV bucket.
func (c *ChattoCore) GetKVKeys(ctx context.Context, bucketName string) ([]string, error) {
	kv, err := c.js.KeyValue(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	keyLister, err := kv.ListKeys(ctx, jetstream.MetaOnly())
	if err != nil {
		// If bucket is empty, ListKeys returns an error
		if err == jetstream.ErrNoKeysFound {
			return []string{}, nil
		}
		return nil, err
	}

	keys := []string{}
	for key := range keyLister.Keys() {
		keys = append(keys, key)
	}

	// Sort for consistent ordering
	sort.Strings(keys)

	return keys, nil
}
