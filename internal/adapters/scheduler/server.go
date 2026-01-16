package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/linkflow-ai/linkflow/internal/core/domain/schedule"
)

type Server struct {
	poller     *Poller
	dispatcher *Dispatcher
	leader     *LeaderElection
	metrics    *Metrics
	config     Config
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

type Config struct {
	PollInterval   time.Duration
	DispatchBuffer int
	LeaderLease    time.Duration
	InstanceID     string
}

func DefaultConfig() Config {
	return Config{
		PollInterval:   time.Minute,
		DispatchBuffer: 100,
		LeaderLease:    30 * time.Second,
		InstanceID:     "",
	}
}

func NewServer(
	scheduleRepo schedule.Repository,
	dispatcher *Dispatcher,
	leader *LeaderElection,
	config Config,
) *Server {
	return &Server{
		poller:     NewPoller(scheduleRepo, config.PollInterval),
		dispatcher: dispatcher,
		leader:     leader,
		metrics:    NewMetrics(),
		config:     config,
		stopCh:     make(chan struct{}),
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Start leader election
	if s.leader != nil {
		go s.leader.Run(ctx)
	}

	// Start the main loop
	s.wg.Add(1)
	go s.run(ctx)

	return nil
}

func (s *Server) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Only poll if we're the leader
			if s.leader != nil && !s.leader.IsLeader() {
				continue
			}

			s.pollAndDispatch(ctx)
		}
	}
}

func (s *Server) pollAndDispatch(ctx context.Context) {
	start := time.Now()

	schedules, err := s.poller.Poll(ctx)
	if err != nil {
		s.metrics.RecordPollError()
		return
	}

	s.metrics.RecordPollDuration(time.Since(start))
	s.metrics.RecordSchedulesFound(len(schedules))

	for _, sched := range schedules {
		if err := s.dispatcher.Dispatch(ctx, sched); err != nil {
			s.metrics.RecordDispatchError()
			continue
		}
		s.metrics.RecordDispatch()
	}
}

func (s *Server) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Server) IsLeader() bool {
	if s.leader == nil {
		return true
	}
	return s.leader.IsLeader()
}

func (s *Server) Metrics() *Metrics {
	return s.metrics
}
