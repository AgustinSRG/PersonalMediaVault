// Task manager

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

// Task manager status data
type TaskManager struct {
	vault *Vault // Reference to the vault

	lock *sync.Mutex // Lock to control access

	pending_tasks      PendingTasksData // Pending tasks data
	pending_tasks_file string           // File to store pending tasks status

	tasks map[uint64]*ActiveTask // Active tasks
	queue []*ActiveTask          // List of tasks waiting to start

	running_count int32 // Counter of running tasks
	max_tasks     int32 // Max number of parallel tasks
}

// Active task status data
type ActiveTask struct {
	definition      *TaskDefinition // Task definition
	running         bool            // True if running
	waiting_session bool            // True if the task needs credentials
	killed          bool            // True if killed

	session *ActiveSession // Reference to the associated session
	status  *TaskStatus    // Task status
}

// Task status data
type TaskStatus struct {
	Stage      string  `json:"stage"`          // Name of the stage
	StageStart int64   `json:"stage_start"`    // Timestamp (Unix millis) of stage start
	Progress   float64 `json:"stage_progress"` // Stage progress (0-100)

	lock *sync.Mutex // Lock to control acess to status data
}

// Get task status
// Returns (1) Stage name
// Returns (2) Timestamp (Unix millis) of stage start
// Returns (3) Stage progress (0-100)
func (s *TaskStatus) Get() (string, int64, float64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.Stage, s.StageStart, s.Progress
}

// Sets stage name
// Auto sets stage start
// Resets progress to 0
// stage - Stage name
func (s *TaskStatus) SetStage(stage string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Stage = stage
	s.StageStart = time.Now().UnixMilli()
	s.Progress = 0
}

// Sets stage progress
// p - Progress (0-100)
func (s *TaskStatus) SetProgress(p float64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Progress = p
}

type TaskDefinitionType uint16

const (
	TASK_ENCODE_ORIGINAL   TaskDefinitionType = 0 // Encoding original asset
	TASK_ENCODE_RESOLUTION TaskDefinitionType = 1 // Encoding extra resolution
	TASK_IMAGE_PREVIEWS    TaskDefinitionType = 2 // making previews images for videos
)

// Task definition data
type TaskDefinition struct {
	Id         uint64                `json:"id"`         // Task ID
	MediaId    uint64                `json:"media_id"`   // Media file ID
	Type       TaskDefinitionType    `json:"type"`       // Task type
	Resolution *UserConfigResolution `json:"resolution"` // Resolution data
}

// Pending tasks data
type PendingTasksData struct {
	NextId  uint64                     `json:"next_id"` // ID for the next task
	Pending map[uint64]*TaskDefinition `json:"pending"` // Pending tasks
}

// Initializes task manager
// base_path - Vault path
// vault - Reference to the vault
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
					lock:       &sync.Mutex{},
				},
				running:         false,
				waiting_session: true,
				session:         nil,
				killed:          false,
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

// Loads configuration
// key - Vault decryption key
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

// Save pending tasks data
func (tm *TaskManager) SavePendingTasks() error {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	// Get the json data
	jsonData, err := json.Marshal(tm.pending_tasks)

	if err != nil {
		return err
	}

	// Make a temp file
	tFile := GetTemporalFileName("json", true)

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

// Call when a new session is created
// Provides credentials to tasks that need them
// session - Session reference
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

	if err != nil {
		return err
	}

	tm.RunPendingTasks()

	return nil
}

// Runs a task
// task - The task
func (tm *TaskManager) RunTask(task *ActiveTask) {
	task.Run(tm.vault) // Run task

	// After task has ended, remove it

	tm.lock.Lock()

	delete(tm.tasks, task.definition.Id)
	delete(tm.pending_tasks.Pending, task.definition.Id)
	tm.running_count--

	tm.lock.Unlock()

	// Save
	err := tm.SavePendingTasks()

	if err != nil {
		LogError(err)
	}

	// Run other tasks
	tm.RunPendingTasks()
}

// Runs pending tasks if possible
func (tm *TaskManager) RunPendingTasks() {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	// Pre-sort queue
	sort.Slice(tm.queue, func(i, j int) bool {
		if tm.queue[i].definition.Type < tm.queue[j].definition.Type {
			return true
		} else if tm.queue[i].definition.Id < tm.queue[j].definition.Id {
			return true
		} else {
			return false
		}
	})

	for len(tm.queue) > 0 && (tm.max_tasks <= 0 || tm.running_count < tm.max_tasks) {
		// Spawn next task

		nextTask := tm.queue[0]

		nextTask.running = true
		tm.queue = tm.queue[1:] // Remove from queue

		go tm.RunTask(nextTask) // Run

		tm.running_count++
	}
}

// Creates a task
// session - Session that cretaes the task
// media_id - Media file ID
// task_type - Task type
// resolution - Resolution data (if task requires it)
// Returns the Id of the new task
func (tm *TaskManager) AddTask(session *ActiveSession, media_id uint64, task_type TaskDefinitionType, resolution *UserConfigResolution) uint64 {
	tm.lock.Lock()

	tm.pending_tasks.NextId++
	task_id := tm.pending_tasks.NextId

	task_definition := TaskDefinition{
		Id:         task_id,
		MediaId:    media_id,
		Type:       task_type,
		Resolution: resolution,
	}

	task := ActiveTask{
		definition: &task_definition,
		status: &TaskStatus{
			Stage:      "",
			StageStart: 0,
			Progress:   0,
			lock:       &sync.Mutex{},
		},
		running:         false,
		waiting_session: false,
		session:         session,
		killed:          false,
	}

	tm.tasks[task_id] = &task

	tm.pending_tasks.Pending[task_id] = task.definition // Add to pending list

	tm.queue = append(tm.queue, &task) // Enqueue

	tm.lock.Unlock()

	// Save
	err := tm.SavePendingTasks()

	if err != nil {
		LogError(err)
	}

	// Run other tasks
	tm.RunPendingTasks()

	return task_id
}

// Kills a task
// task_id - ID of the task
func (tm *TaskManager) KillTask(task_id uint64) {
	tm.lock.Lock()

	if tm.tasks[task_id] == nil {
		tm.lock.Unlock()
		return
	}

	tm.tasks[task_id].killed = true
	delete(tm.pending_tasks.Pending, task_id) // Remove from pending tasks list

	tm.lock.Unlock()

	// Save
	err := tm.SavePendingTasks()

	if err != nil {
		LogError(err)
	}

	// Run other tasks
	tm.RunPendingTasks()
}

// Kill every task given a media ID
// media_id - Media file ID
func (tm *TaskManager) KillTaskByMedia(media_id uint64) {
	tm.lock.Lock()

	for task_id, task := range tm.tasks {
		if task.definition.MediaId == media_id {
			tm.tasks[task_id].killed = true
			delete(tm.pending_tasks.Pending, task_id)
		}
	}

	tm.lock.Unlock()

	// Save
	err := tm.SavePendingTasks()

	if err != nil {
		LogError(err)
	}

	// Run other tasks
	tm.RunPendingTasks()
}

// Get task status
// task_id - ID of the task
// Returns task status
func (tm *TaskManager) GetTaskStatus(task_id uint64) *TaskStatus {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if tm.tasks[task_id] == nil {
		return nil
	}

	return tm.tasks[task_id].status
}

// Task list info data struct for API
type TaskListInfoEntry struct {
	Id      uint64 `json:"id"`      // Task ID
	Running bool   `json:"running"` // True if running

	MediaId    uint64                `json:"media_id"`   // Media file ID
	Type       TaskDefinitionType    `json:"type"`       // Task type
	Resolution *UserConfigResolution `json:"resolution"` // Resolution data

	Stage      string  `json:"stage"`          // Name of current stage
	StageStart int64   `json:"stage_start"`    // Stage start timestamp (unix millis)
	Now        int64   `json:"time_now"`       // Server time (unix millis)
	Progress   float64 `json:"stage_progress"` // Stage progress (0-100)
}

// Gets task status info for API
// task_id - Task ID
// Returns task info
func (tm *TaskManager) GetTaskInfo(task_id uint64) *TaskListInfoEntry {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if tm.tasks[task_id] == nil {
		return nil
	}

	task := tm.tasks[task_id]

	var info TaskListInfoEntry

	info.Id = task.definition.Id
	info.Running = task.running

	info.MediaId = task.definition.MediaId
	info.Type = task.definition.Type
	info.Resolution = task.definition.Resolution

	stage, stage_start, stage_p := task.status.Get()

	info.Stage = stage
	info.StageStart = stage_start
	info.Now = time.Now().UnixMilli()
	info.Progress = stage_p

	return &info
}

// Gets information of all active tasks for API
func (tm *TaskManager) GetAllTasks() []*TaskListInfoEntry {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	result := make([]*TaskListInfoEntry, 0)

	for _, task := range tm.tasks {
		var info TaskListInfoEntry

		info.Id = task.definition.Id
		info.Running = task.running

		info.MediaId = task.definition.MediaId
		info.Type = task.definition.Type
		info.Resolution = task.definition.Resolution

		stage, stage_start, stage_p := task.status.Get()

		info.Stage = stage
		info.StageStart = stage_start
		info.Now = time.Now().UnixMilli()
		info.Progress = stage_p

		result = append(result, &info)
	}

	return result
}
