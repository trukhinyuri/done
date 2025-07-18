package bolt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"time"

	database "done/lib/database/interface"
	"done/lib/utils"
)

const (
	tasksBucket          = "tasks"
	completedTasksBucket = "tasks_completed"
	gamificationBucket   = "gamification"
	gamificationKey      = "stats"
)

type BoltDB struct {
	db     *bolt.DB
	dbPath string
}

func NewBoltDB(dbPath string) *BoltDB {
	return &BoltDB{
		dbPath: dbPath,
	}
}

func (b *BoltDB) Connect() error {
	db, err := bolt.Open(b.dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return fmt.Errorf("failed to open bolt database: %w", err)
	}

	b.db = db

	// Create buckets if they don't exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(tasksBucket))
		if err != nil {
			return fmt.Errorf("failed to create tasks bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(completedTasksBucket))
		if err != nil {
			return fmt.Errorf("failed to create completed tasks bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(gamificationBucket))
		if err != nil {
			return fmt.Errorf("failed to create gamification bucket: %w", err)
		}

		return nil
	})

	return err
}

func (b *BoltDB) Disconnect() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

func (b *BoltDB) GetTasks() ([]database.Task, error) {
	var tasks []database.Task

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tasksBucket))
		if bucket == nil {
			return errors.New("tasks bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var task database.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}
			// Clean task body to remove any encoding artifacts
			task.Body = utils.CleanTaskText(task.Body)
			tasks = append(tasks, task)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	// Sort tasks by Order field
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].Order > tasks[j].Order {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	return tasks, nil
}

func (b *BoltDB) GetTaskByUUID(uuid string) (*database.Task, error) {
	var task database.Task

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tasksBucket))
		if bucket == nil {
			return errors.New("tasks bucket not found")
		}

		data := bucket.Get([]byte(uuid))
		if data == nil {
			return errors.New("task not found")
		}

		return json.Unmarshal(data, &task)
	})

	if err != nil {
		return nil, err
	}

	// Clean task body to remove any encoding artifacts
	task.Body = utils.CleanTaskText(task.Body)

	return &task, nil
}

func (b *BoltDB) AddTask(task *database.Task) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tasksBucket))
		if bucket == nil {
			return errors.New("tasks bucket not found")
		}

		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(task.UUID), data)
	})
}

func (b *BoltDB) UpdateTask(task *database.Task) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tasksBucket))
		if bucket == nil {
			return errors.New("tasks bucket not found")
		}

		// Check if task exists
		if bucket.Get([]byte(task.UUID)) == nil {
			return errors.New("task not found")
		}

		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(task.UUID), data)
	})
}

func (b *BoltDB) RemoveTask(uuid string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(tasksBucket))
		if bucket == nil {
			return errors.New("tasks bucket not found")
		}

		return bucket.Delete([]byte(uuid))
	})
}

func (b *BoltDB) GetCompletedTasks() ([]database.Task, error) {
	var tasks []database.Task

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(completedTasksBucket))
		if bucket == nil {
			return errors.New("completed tasks bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var task database.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}
			// Clean task body to remove any encoding artifacts
			task.Body = utils.CleanTaskText(task.Body)
			tasks = append(tasks, task)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (b *BoltDB) AddCompletedTask(task *database.Task) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(completedTasksBucket))
		if bucket == nil {
			return errors.New("completed tasks bucket not found")
		}

		data, err := json.Marshal(task)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(task.UUID), data)
	})
}

func (b *BoltDB) GetGamification() (*database.Gamification, error) {
	var gamification database.Gamification

	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(gamificationBucket))
		if bucket == nil {
			return errors.New("gamification bucket not found")
		}

		data := bucket.Get([]byte(gamificationKey))
		if data == nil {
			// Return default gamification data if none exists
			return nil
		}

		return json.Unmarshal(data, &gamification)
	})

	if err != nil {
		return nil, err
	}

	// If no data exists, return initialized gamification
	if gamification.Level == 0 {
		gamification.Level = 1
	}

	return &gamification, nil
}

func (b *BoltDB) UpdateGamification(gamification *database.Gamification) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(gamificationBucket))
		if bucket == nil {
			return errors.New("gamification bucket not found")
		}

		data, err := json.Marshal(gamification)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(gamificationKey), data)
	})
}

func (b *BoltDB) DBUpgrade() string {
	return "DBUpgrade not required for BoltDB"
}

