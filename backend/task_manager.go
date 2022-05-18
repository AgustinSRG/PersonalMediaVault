// Task manager

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type TaskManager struct {
	vault *Vault

	lock *sync.Mutex

	pending_tasks      PendingTasksData
	pending_tasks_file string

	tasks map[uint64]*ActiveTask
	queue []*ActiveTask

	running_count int32
	max_tasks     int32
}

type ActiveTask struct {
	definition      *TaskDefinition
	running         bool
	waiting_session bool // True if the task needs credentials

	session *ActiveSession
	status  *TaskStatus
}

type TaskStatus struct {
	Stage      string  `json:"stage"`
	StageStart int64   `json:"stage_start"`
	Progress   float64 `json:"stage_progress"`
}

type TaskDefinition struct {
	Id                    uint64 `json:"id"`
	MediaId               uint64 `json:"media_id"`
	UseOriginalResolution bool   `json:"original_resolution"`
	Width                 int32  `json:"width"`
	Height                int32  `json:"height"`
	Fps                   int32  `json:"fps"`
}

type PendingTasksData struct {
	NextId  uint64                     `json:"next_id"`
	Pending map[uint64]*TaskDefinition `json:"pending"`
}

func (tm *TaskManager) Initialize(base_path string, vault *Vault) error {
	tm.vault = vault
	tm.lock = &sync.Mutex{}

	tm.running_count = 0
	tm.max_tasks = 1
	tm.tasks = make(map[uint64]*ActiveTask)
	tm.queue = make([]*ActiveTask, 0)

	file := path.Join(base_path, "tasks.json")
	tm.pending_tasks_file = file

	if _, err := os.Stat(file); err == nil {
		// exists
		b, err := ioutil.ReadFile(file)

		if err != nil {
			return err
		}

		// Parse
		err = json.Unmarshal(b, &tm.pending_tasks)

		if err != nil {
			return err
		}

		// Initialize pending tasks

		if tm.pending_tasks.Pending == nil {
			tm.pending_tasks.Pending = make(map[uint64]*TaskDefinition)
		}

		for task_id, task_definition := range tm.pending_tasks.Pending {
			if task_definition == nil {
				continue
			}
			task := ActiveTask{
				definition: task_definition,
				status: &TaskStatus{
					Stage:      "",
					StageStart: 0,
					Progress:   0,
				},
				running:         false,
				waiting_session: true,
				session:         nil,
			}

			tm.tasks[task_id] = &task
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		tm.pending_tasks.NextId = 0
		tm.pending_tasks.Pending = make(map[uint64]*TaskDefinition)
	} else {
		return err
	}

	return nil
}

func (tm *TaskManager) LoadUserConfigParams(key []byte) error {
	uc, err := tm.vault.config.Read(key)

	if err != nil {
		return err
	}

	tm.lock.Lock()
	defer tm.lock.Unlock()

	tm.max_tasks = uc.MaxTasks

	return nil
}

func (tm *TaskManager) SavePendingTasks() error {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	// Get the json data
	jsonData, err := json.Marshal(tm.pending_tasks)

	if err != nil {
		return err
	}

	// Make a temp file
	tFile := GetTemporalFileName("json")

	// Write file
	err = ioutil.WriteFile(tFile, jsonData, FILE_PERMISSION)
	if err != nil {
		return err
	}

	// Move to the original path
	err = os.Rename(tFile, tm.pending_tasks_file)
	if err != nil {
		return err
	}

	return nil
}

func (tm *TaskManager) OnNewSession(session *ActiveSession) error {
	tm.lock.Lock()

	// Check for tasks waiting for a session and queue them
	for _, task := range tm.tasks {
		if task.waiting_session {
			task.session = session
			task.waiting_session = false
			tm.queue = append(tm.queue, task)
		}
	}

	tm.lock.Unlock()

	// Update user config

	err := tm.LoadUserConfigParams(session.key)

	return err
}
