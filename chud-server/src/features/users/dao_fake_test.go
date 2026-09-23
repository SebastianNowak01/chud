package users

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"time"

	"github.com/sebnow/chud/platform/db"
)

type fakeUserDAO struct {
	mu    sync.Mutex
	users map[string]User
}

func newFakeUserDAO() *fakeUserDAO {
	return &fakeUserDAO{users: make(map[string]User)}
}

func (d *fakeUserDAO) GetAllUsers(_ context.Context) ([]User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := make([]User, 0, len(d.users))
	for _, u := range d.users {
		result = append(result, u)
	}
	slices.SortFunc(result, func(a, b User) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), cmp.Compare(a.Username, b.Username))
	})
	return result, nil
}

func (d *fakeUserDAO) GetUserByID(_ context.Context, id string) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	u, ok := d.users[id]
	if !ok {
		return nil, db.ErrNotFound
	}
	return &u, nil
}

func (d *fakeUserDAO) GetUserByUsername(_ context.Context, username string) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, u := range d.users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, db.ErrNotFound
}

func (d *fakeUserDAO) InsertUser(_ context.Context, user *User) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.usernameTaken(user.Username, user.ID) {
		return nil, db.ErrAlreadyExists
	}
	u := *user
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	d.users[u.ID] = u
	return &u, nil
}

func (d *fakeUserDAO) UpdateUser(_ context.Context, user *User) (*User, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	existing, ok := d.users[user.ID]
	if !ok {
		return nil, db.ErrNotFound
	}
	if d.usernameTaken(user.Username, user.ID) {
		return nil, db.ErrAlreadyExists
	}
	u := *user
	u.CreatedAt = existing.CreatedAt
	u.UpdatedAt = time.Now()
	d.users[u.ID] = u
	return &u, nil
}

func (d *fakeUserDAO) DeleteUser(_ context.Context, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.users[id]; !ok {
		return db.ErrNotFound
	}
	delete(d.users, id)
	return nil
}

func (d *fakeUserDAO) DemoteAdminsExcept(_ context.Context, username string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, u := range d.users {
		if u.IsAdmin && u.Username != username {
			u.IsAdmin = false
			d.users[id] = u
		}
	}
	return nil
}

func (d *fakeUserDAO) usernameTaken(username, exceptID string) bool {
	for id, u := range d.users {
		if id != exceptID && u.Username == username {
			return true
		}
	}
	return false
}
