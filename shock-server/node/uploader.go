package node

import (
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
)

var (
	UploadQueue   chan string
	uploadManager *UploadManager
)

type UploadManager struct {
	workers int
}

// InitUploader initializes the upload worker pool
func InitUploader() {
	if !conf.AUTO_UPLOAD || conf.DEFAULT_LOCATION == "" {
		logger.Info("(InitUploader) Auto-upload disabled or no default location configured")
		return
	}

	// Validate that the default location exists
	if _, ok := conf.LocationsMap[conf.DEFAULT_LOCATION]; !ok {
		logger.Errorf("(InitUploader) Default location '%s' not found in Locations.yaml", conf.DEFAULT_LOCATION)
		return
	}

	workers := conf.UPLOAD_WORKERS
	if workers <= 0 {
		workers = 4
	}

	// Create buffered channel for upload queue
	UploadQueue = make(chan string, 1000)

	uploadManager = &UploadManager{
		workers: workers,
	}

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		go uploadManager.worker(i)
	}

	logger.Infof("(InitUploader) UploadManager started with %d workers, target location: %s", workers, conf.DEFAULT_LOCATION)
}

// QueueUpload adds a node ID to the upload queue (non-blocking)
func QueueUpload(nodeId string) {
	if UploadQueue == nil {
		return
	}

	select {
	case UploadQueue <- nodeId:
		logger.Debug(2, "(QueueUpload) Queued node %s for upload", nodeId)
	default:
		logger.Errorf("(QueueUpload) Upload queue full, dropping node %s", nodeId)
	}
}

// worker processes uploads from the queue
func (um *UploadManager) worker(id int) {
	logger.Debug(1, "(UploadManager) Worker %d started", id)

	for nodeId := range UploadQueue {
		um.processUpload(nodeId, id)
	}
}

// processUpload handles the actual upload of a node to S3
func (um *UploadManager) processUpload(nodeId string, workerId int) {
	logger.Debug(2, "(UploadManager) Worker %d processing node %s", workerId, nodeId)

	// Lock the node
	err := locker.NodeLockMgr.LockNode(nodeId)
	if err != nil {
		logger.Errorf("(UploadManager) Failed to lock node %s: %s", nodeId, err.Error())
		return
	}
	defer locker.NodeLockMgr.UnlockNode(nodeId)

	// Load the node
	node, err := Load(nodeId)
	if err != nil {
		logger.Errorf("(UploadManager) Failed to load node %s: %s", nodeId, err.Error())
		return
	}

	// Check if file exists
	if node.File.Size == 0 {
		logger.Debug(2, "(UploadManager) Node %s has no file, skipping", nodeId)
		return
	}

	// Check if already at the target location
	for _, loc := range node.Locations {
		if loc.ID == conf.DEFAULT_LOCATION && loc.Stored {
			logger.Debug(2, "(UploadManager) Node %s already stored at %s, skipping", nodeId, conf.DEFAULT_LOCATION)
			return
		}
	}

	// Get location config
	locationConfig, ok := conf.LocationsMap[conf.DEFAULT_LOCATION]
	if !ok {
		logger.Errorf("(UploadManager) Location %s not found", conf.DEFAULT_LOCATION)
		return
	}

	// Only handle S3 type locations for now
	if locationConfig.Type != "S3" {
		logger.Errorf("(UploadManager) Location type %s not supported for auto-upload", locationConfig.Type)
		return
	}

	// Perform the upload
	err, verified := S3Upload(nodeId, node, locationConfig)
	if err != nil {
		logger.Errorf("(UploadManager) Failed to upload node %s to %s: %s", nodeId, conf.DEFAULT_LOCATION, err.Error())
		return
	}

	if !verified {
		logger.Errorf("(UploadManager) Upload verification failed for node %s", nodeId)
		return
	}

	// Add location to node
	now := time.Now()
	newLocation := Location{
		ID:            conf.DEFAULT_LOCATION,
		Stored:        true,
		RequestedDate: &now,
	}

	err = node.AddLocation(newLocation)
	if err != nil {
		// Location might already exist (race condition), try to update it
		logger.Debug(2, "(UploadManager) AddLocation for node %s returned: %s", nodeId, err.Error())
	}

	// Save the node
	err = node.Save()
	if err != nil {
		logger.Errorf("(UploadManager) Failed to save node %s: %s", nodeId, err.Error())
		return
	}

	logger.Infof("(UploadManager) Successfully uploaded node %s to %s", nodeId, conf.DEFAULT_LOCATION)
}
