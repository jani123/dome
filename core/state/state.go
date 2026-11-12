// Package state owns durable execution data. Stored values must never contain secrets.
package state

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/resource"
	"golang.org/x/sys/unix"
	_ "modernc.org/sqlite"
)

type (
	Store struct {
		db    *sql.DB
		path  string
		guard sync.Mutex
	}
	Managed struct {
		ID       resource.ID       `json:"id"`
		Applied  resource.Snapshot `json:"applied"`
		Revision string            `json:"revision"`
	}
	Operation struct {
		ID       string    `json:"id"`
		Resource string    `json:"resource"`
		Action   string    `json:"action"`
		Status   string    `json:"status"`
		Code     string    `json:"error_code,omitempty"`
		Started  string    `json:"started,omitempty"`
		Nodes    []Node    `json:"nodes,omitempty"`
		Attempts []Attempt `json:"attempts,omitempty"`
	}
	Execution struct {
		ID          string `json:"id"`
		Playbook    string `json:"playbook"`
		Fingerprint string `json:"fingerprint"`
		Status      string `json:"status"`
	}
	Node struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Attempts int    `json:"attempts"`
		Code     string `json:"error_code,omitempty"`
	}
)

func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func Open(path string) (_ *Store, err error) {
	defer func() { err = stateError(err) }()
	if path == "" {
		return nil, fault.New(fault.Validation, "SQLite path is required")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			return nil, err
		}
		f.Close()
		if err = os.Chmod(path, 0o600); err != nil {
			return nil, err
		}
	}
	if path != ":memory:" {
		absolute, resolveErr := filepath.Abs(path)
		if resolveErr != nil {
			return nil, resolveErr
		}
		path, resolveErr = filepath.EvalSymlinks(absolute)
		if resolveErr != nil {
			return nil, resolveErr
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, path: path}
	for _, stmt := range []string{"PRAGMA foreign_keys=ON", "PRAGMA busy_timeout=5000", "PRAGMA journal_mode=WAL"} {
		if _, err = db.Exec(stmt); err != nil {
			db.Close()
			return nil, err
		}
	}
	if err = s.migrate(); err != nil {
		db.Close()
		return nil, fault.Wrap(fault.Persistence, err)
	}
	return s, nil
}

func (s *Store) Close() (err error) {
	defer func() { err = stateError(err) }()
	return s.db.Close()
}

func (s *Store) migrate() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY);`); err != nil {
		return err
	}
	var version int
	if err = tx.QueryRow("SELECT COALESCE(MAX(version),0) FROM schema_migrations").Scan(&version); err != nil {
		return err
	}
	if version > 2 {
		return fmt.Errorf("database schema is newer than this binary")
	}
	if version == 0 {
		_, err = tx.Exec(`
 CREATE TABLE managed(module TEXT NOT NULL,name TEXT NOT NULL,exists_flag INTEGER NOT NULL,value TEXT NOT NULL,revision TEXT NOT NULL,PRIMARY KEY(module,name));
 CREATE TABLE metadata(key TEXT PRIMARY KEY,value TEXT NOT NULL);
 CREATE TABLE operations(id TEXT PRIMARY KEY,resource TEXT NOT NULL,action TEXT NOT NULL,status TEXT NOT NULL,error_code TEXT NOT NULL DEFAULT '',started TEXT NOT NULL);
 CREATE TABLE executions(id TEXT PRIMARY KEY,playbook TEXT NOT NULL,fingerprint TEXT NOT NULL,status TEXT NOT NULL);
 CREATE TABLE nodes(execution_id TEXT NOT NULL REFERENCES executions(id) ON DELETE CASCADE,id TEXT NOT NULL,status TEXT NOT NULL,attempts INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(execution_id,id));
 CREATE TABLE locks(name TEXT PRIMARY KEY,owner TEXT NOT NULL);
 INSERT INTO schema_migrations VALUES(1);`)
		if err != nil {
			return err
		}
	}
	if version < 2 {
		if _, err = tx.Exec(`ALTER TABLE nodes ADD COLUMN error_code TEXT NOT NULL DEFAULT '';
CREATE TABLE attempts(execution_id TEXT NOT NULL,node_id TEXT NOT NULL,number INTEGER NOT NULL,status TEXT NOT NULL,error_code TEXT NOT NULL,started TEXT NOT NULL,finished TEXT NOT NULL,PRIMARY KEY(execution_id,node_id,number),FOREIGN KEY(execution_id,node_id) REFERENCES nodes(execution_id,id) ON DELETE CASCADE);
INSERT INTO schema_migrations VALUES(2);`); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Guard provides a process lifetime lock. The kernel releases it after a crash.
// All operations against one database must use this guard, including playbooks.
func (s *Store) Guard(ctx context.Context) (_ func(), err error) {
	defer func() { err = stateError(err) }()
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if !s.guard.TryLock() {
		return nil, fault.New(fault.Conflict, "another operation is using this state store")
	}
	if s.path == ":memory:" {
		return s.guard.Unlock, nil
	}
	f, err := os.OpenFile(s.path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		s.guard.Unlock()
		return nil, err
	}
	if err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		s.guard.Unlock()
		return nil, fault.New(fault.Conflict, "another agent is using this state database")
	}
	if err = ctx.Err(); err != nil {
		f.Close()
		s.guard.Unlock()
		return nil, err
	}
	// No other process can own live locks once the exclusive guard is held.
	if _, err = s.db.ExecContext(ctx, "DELETE FROM locks"); err != nil {
		f.Close()
		s.guard.Unlock()
		return nil, err
	}
	return func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN); _ = f.Close(); s.guard.Unlock() }, nil
}

func (s *Store) Managed(ctx context.Context) (_ []Managed, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT module,name,exists_flag,value,revision FROM managed ORDER BY module,name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Managed{}
	for rows.Next() {
		var v Managed
		if err = rows.Scan(&v.ID.Module, &v.ID.Name, &v.Applied.Exists, &v.Applied.Value, &v.Revision); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) SaveManaged(ctx context.Context, v Managed) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, `INSERT INTO managed VALUES(?,?,?,?,?) ON CONFLICT(module,name) DO UPDATE SET exists_flag=excluded.exists_flag,value=excluded.value,revision=excluded.revision`, v.ID.Module, v.ID.Name, v.Applied.Exists, v.Applied.Value, v.Revision)
	return err
}

func (s *Store) Forget(ctx context.Context, id resource.ID) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "DELETE FROM managed WHERE module=? AND name=?", id.Module, id.Name)
	return err
}

func (s *Store) Revision(ctx context.Context) (_ string, err error) {
	defer func() { err = stateError(err) }()
	var v string
	err = s.db.QueryRowContext(ctx, "SELECT value FROM metadata WHERE key='revision'").Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

func (s *Store) SetRevision(ctx context.Context, v string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "INSERT INTO metadata VALUES('revision',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", v)
	return err
}

func (s *Store) BeginOperation(ctx context.Context, r, action string) (_ string, err error) {
	defer func() { err = stateError(err) }()
	id := NewID()
	_, err = s.db.ExecContext(ctx, "INSERT INTO operations(id,resource,action,status,started) VALUES(?,?,?,'running',?)", id, r, action, time.Now().UTC().Format(time.RFC3339Nano))
	return id, err
}

func (s *Store) EndOperation(ctx context.Context, id, status, code string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "UPDATE operations SET status=?,error_code=? WHERE id=?", status, code, id)
	return err
}

func (s *Store) Operations(ctx context.Context) (_ []Operation, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT id,resource,action,status,error_code,started FROM operations ORDER BY started,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Operation{}
	for rows.Next() {
		var v Operation
		if err = rows.Scan(&v.ID, &v.Resource, &v.Action, &v.Status, &v.Code, &v.Started); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) NewExecution(ctx context.Context, id, book, fp string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "INSERT INTO executions VALUES(?,?,?,'running')", id, book, fp)
	return err
}

func (s *Store) Execution(ctx context.Context, id string) (_ Execution, err error) {
	defer func() { err = stateError(err) }()
	var v Execution
	err = s.db.QueryRowContext(ctx, "SELECT id,playbook,fingerprint,status FROM executions WHERE id=?", id).Scan(&v.ID, &v.Playbook, &v.Fingerprint, &v.Status)
	return v, err
}

func (s *Store) FinishExecution(ctx context.Context, id, status string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "UPDATE executions SET status=? WHERE id=?", status, id)
	return err
}

func (s *Store) Nodes(ctx context.Context, id string) (_ map[string]Node, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT id,status,attempts,error_code FROM nodes WHERE execution_id=?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]Node{}
	for rows.Next() {
		var v Node
		if err = rows.Scan(&v.ID, &v.Status, &v.Attempts, &v.Code); err != nil {
			return nil, err
		}
		out[v.ID] = v
	}
	return out, rows.Err()
}

func (s *Store) SaveNode(ctx context.Context, execution string, v Node) (err error) {
	defer func() { err = stateError(err) }()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO nodes(execution_id,id,status,attempts,error_code) VALUES(?,?,?,?,?) ON CONFLICT(execution_id,id) DO UPDATE SET status=excluded.status,attempts=excluded.attempts,error_code=excluded.error_code`, execution, v.ID, v.Status, v.Attempts, v.Code); err != nil {
		return err
	}
	if v.Attempts > 0 && (v.Status == "running" || v.Status == "succeeded" || v.Status == "failed") {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		finished := ""
		if v.Status != "running" {
			finished = now
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO attempts VALUES(?,?,?,?,?,?,?) ON CONFLICT(execution_id,node_id,number) DO UPDATE SET status=excluded.status,error_code=excluded.error_code,finished=excluded.finished WHERE attempts.status='running'`, execution, v.ID, v.Attempts, v.Status, v.Code, now, finished); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Acquire(ctx context.Context, names []string, owner string) (_ bool, err error) {
	defer func() { err = stateError(err) }()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	for _, name := range names {
		var existing string
		err = tx.QueryRowContext(ctx, "SELECT owner FROM locks WHERE name=?", name).Scan(&existing)
		if err == nil && existing != owner {
			return false, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return false, err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO locks VALUES(?,?) ON CONFLICT(name) DO NOTHING", name, owner); err != nil {
			return false, err
		}
	}
	return true, tx.Commit()
}

func (s *Store) Release(ctx context.Context, owner string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "DELETE FROM locks WHERE owner=?", owner)
	return err
}

func (s *Store) Metadata(ctx context.Context, key string) (_ string, err error) {
	defer func() { err = stateError(err) }()
	var v string
	err = s.db.QueryRowContext(ctx, "SELECT value FROM metadata WHERE key=?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

func (s *Store) SetMetadata(ctx context.Context, key, value string) (err error) {
	defer func() { err = stateError(err) }()
	_, err = s.db.ExecContext(ctx, "INSERT INTO metadata VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, value)
	return err
}

func (s *Store) Executions(ctx context.Context) (_ []Execution, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT id,playbook,fingerprint,status FROM executions ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Execution{}
	for rows.Next() {
		var v Execution
		if err = rows.Scan(&v.ID, &v.Playbook, &v.Fingerprint, &v.Status); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

type Attempt struct {
	Node     string `json:"node_id"`
	Number   int    `json:"number"`
	Status   string `json:"status"`
	Code     string `json:"error_code,omitempty"`
	Started  string `json:"started"`
	Finished string `json:"finished,omitempty"`
}

func (s *Store) Attempts(ctx context.Context, execution string) (_ []Attempt, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT node_id,number,status,error_code,started,finished FROM attempts WHERE execution_id=? ORDER BY node_id,number", execution)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Attempt{}
	for rows.Next() {
		var v Attempt
		if err = rows.Scan(&v.Node, &v.Number, &v.Status, &v.Code, &v.Started, &v.Finished); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func stateError(err error) error {
	if err == nil {
		return nil
	}
	if fault.Category(err) != "internal" {
		return err
	}
	return fault.Wrap(fault.Persistence, err, "module", "state")
}

// MetadataByPrefix lists durable recovery intents without interpreting their payloads.
func (s *Store) MetadataByPrefix(ctx context.Context, prefix string) (_ map[string]string, err error) {
	defer func() { err = stateError(err) }()
	rows, err := s.db.QueryContext(ctx, "SELECT key,value FROM metadata WHERE substr(key,1,length(?))=? ORDER BY key", prefix, prefix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var key, value string
		if err = rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, rows.Err()
}
