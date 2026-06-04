package events

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// ProjectionManager owns the JetStream consumers for a set of projectors.
// Projectors remain the per-projection state and wait handles; the manager
// groups them by exact canonical subject filters so identical read models
// replay the stream once per process.
type ProjectionManager struct {
	stream     jetstream.Stream
	projectors []*Projector
	logger     Logger
}

// NewProjectionManager returns a manager for projectors bound to the same
// stream. The projectors are not started until Run is called.
func NewProjectionManager(stream jetstream.Stream, projectors []*Projector, logger Logger) *ProjectionManager {
	return &ProjectionManager{
		stream:     stream,
		projectors: append([]*Projector(nil), projectors...),
		logger:     logger,
	}
}

// Run starts one ordered consumer per canonical subject-filter group and
// blocks until ctx is cancelled or one group fails to start.
func (m *ProjectionManager) Run(ctx context.Context) error {
	for _, p := range m.projectors {
		if err := p.start(); err != nil {
			return err
		}
	}

	groups := m.groups()
	if len(groups) == 0 {
		<-ctx.Done()
		return ctx.Err()
	}

	g, gctx := errgroup.WithContext(ctx)
	for _, group := range groups {
		group := group
		g.Go(func() error {
			return group.run(gctx)
		})
	}
	return g.Wait()
}

func (m *ProjectionManager) groups() []*projectionGroup {
	byKey := make(map[string]*projectionGroup)
	keys := make([]string, 0)

	for _, p := range m.projectors {
		subjects := canonicalSubjects(p.Subjects())
		key := strings.Join(subjects, "\x00")
		group, ok := byKey[key]
		if !ok {
			group = &projectionGroup{
				stream:     m.stream,
				subjects:   subjects,
				projectors: nil,
				logger:     m.logger,
			}
			byKey[key] = group
			keys = append(keys, key)
		}
		group.projectors = append(group.projectors, p)
	}

	sort.Strings(keys)
	groups := make([]*projectionGroup, 0, len(keys))
	for _, key := range keys {
		groups = append(groups, byKey[key])
	}
	return groups
}

func canonicalSubjects(subjects []string) []string {
	if len(subjects) == 0 {
		return nil
	}
	out := append([]string(nil), subjects...)
	sort.Strings(out)
	n := 0
	for _, subject := range out {
		if n > 0 && out[n-1] == subject {
			continue
		}
		out[n] = subject
		n++
	}
	return out[:n]
}

type projectionGroup struct {
	stream     jetstream.Stream
	subjects   []string
	projectors []*Projector
	logger     Logger
}

func (g *projectionGroup) run(ctx context.Context) error {
	cons, err := g.stream.OrderedConsumer(ctx, jetstream.OrderedConsumerConfig{
		FilterSubjects:    g.subjects,
		DeliverPolicy:     jetstream.DeliverAllPolicy,
		InactiveThreshold: 30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("create ordered projection consumer for %s: %w", strings.Join(g.subjects, ","), err)
	}

	cc, err := cons.Consume(g.handleMessage,
		jetstream.ConsumeErrHandler(g.handleConsumeErr),
	)
	if err != nil {
		return fmt.Errorf("start projection consumer for %s: %w", strings.Join(g.subjects, ","), err)
	}
	defer cc.Stop()

	<-ctx.Done()
	return ctx.Err()
}

func (g *projectionGroup) handleMessage(msg jetstream.Msg) {
	meta, err := msg.Metadata()
	if err != nil {
		g.logger.Warn("Skipping projection event with no metadata",
			"subject", msg.Subject(),
			"error", err)
		return
	}

	var event corev1.Event
	if err := proto.Unmarshal(msg.Data(), &event); err != nil {
		g.logger.Warn("Skipping unmarshalable projection event",
			"subject", msg.Subject(),
			"seq", meta.Sequence.Stream,
			"error", err)
		for _, p := range g.projectors {
			p.advanceIfHealthy(meta.Sequence.Stream)
		}
		return
	}

	for _, p := range g.projectors {
		p.applyEvent(&event, meta.Sequence.Stream, msg.Subject())
	}
}

func (g *projectionGroup) handleConsumeErr(_ jetstream.ConsumeContext, err error) {
	g.logger.Warn("Projection consumer error (auto-recovering)",
		"subjects", strings.Join(g.subjects, ","),
		"error", err)
}
