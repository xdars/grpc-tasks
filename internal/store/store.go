package store

import (
	"sync"

	pb "github.com/xdars/grpc-tasks/gen/pb"
)

type UserRecord struct {
	ID           string
	Username     string
	PasswordHash string
}

type Store struct {
	mu       sync.RWMutex
	users    map[string]*UserRecord // key: username
	tasks    map[string]*pb.Task    // key: task id
	eventCh  chan *pb.TaskEvent
	watchers map[string]chan *pb.TaskEvent
	watchMu  sync.Mutex
}

func NewStore() *Store {
	s := &Store{
		users:    make(map[string]*UserRecord),
		tasks:    make(map[string]*pb.Task),
		eventCh:  make(chan *pb.TaskEvent, 64),
		watchers: make(map[string]chan *pb.TaskEvent),
	}
	go s.broadcast()
	return s
}

func (s *Store) broadcast() {
	for event := range s.eventCh {
		s.watchMu.Lock()
		for _, ch := range s.watchers {
			select {
			case ch <- event:
			default:
			}
		}
		s.watchMu.Unlock()
	}
}

func (s *Store) Subscribe(userID string) chan *pb.TaskEvent {
	ch := make(chan *pb.TaskEvent, 16)
	s.watchMu.Lock()
	s.watchers[userID] = ch
	s.watchMu.Unlock()
	return ch
}

func (s *Store) Unsubscribe(userID string) {
	s.watchMu.Lock()
	delete(s.watchers, userID)
	s.watchMu.Unlock()
}

func (s *Store) Publish(event *pb.TaskEvent) {
	select {
	case s.eventCh <- event:
	default:
	}
}

func (s *Store) AddUser(u *UserRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[u.Username] = u
}

func (s *Store) GetUser(username string) (*UserRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, exists := s.users[username]
	return u, exists
}

func (s *Store) UserExists(username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.users[username]
	return exists
}

func (s *Store) SaveTask(t *pb.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.Id] = t
}

func (s *Store) GetTask(id string) (*pb.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Store) GetTasksByUser(userID string) []*pb.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*pb.Task
	for _, t := range s.tasks {
		if t.UserId == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) DeleteTask(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}
