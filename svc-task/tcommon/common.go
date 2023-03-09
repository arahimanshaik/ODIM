package tcommon

import (
	"net/http"

	dmtf "github.com/ODIM-Project/ODIM/lib-dmtf/model"
	"github.com/ODIM-Project/ODIM/lib-utilities/common"
	"github.com/ODIM-Project/ODIM/lib-utilities/config"
	l "github.com/ODIM-Project/ODIM/lib-utilities/logs"
	"github.com/ODIM-Project/ODIM/lib-utilities/response"
)

var (
	// ConfigFilePath holds the value of odim config file path
	ConfigFilePath string
)

const (
	// IterationCount is a value that needs to be added in context
	// to track the number of threads created
	IterationCount = "IterationCount"
)

// TaskStatusMap is used to get task state
var TaskStatusMap = map[string]dmtf.TaskState{
	"TaskStarted":          dmtf.TaskStateStarting,
	"TaskProgressChanged":  dmtf.TaskStateRunning,
	"TaskPaused":           dmtf.TaskStateSuspended,
	"TaskAborted":          dmtf.TaskStateInterrupted,
	"TaskCompletedOK":      dmtf.TaskStateCompleted,
	"TaskRemoved":          dmtf.TaskStateKilled,
	"TaskCompletedWarning": dmtf.TaskStateException,
	"TaskCancelled":        dmtf.TaskStateCancelled,
}

// GetStatusCode return status code for the response based on task state and task status
func GetStatusCode(taskState dmtf.TaskState, taskStatus string) int {
	if taskState == dmtf.TaskStateCompleted && taskStatus == common.OK {
		return http.StatusOK
	} else if taskState == dmtf.TaskStateCancelled && taskStatus == common.Critical {
		return http.StatusInternalServerError
	}
	return http.StatusAccepted
}

// GetTaskResponse return status task response using status code and message
func GetTaskResponse(statusCode int, message string) response.RPC {
	var resp response.RPC
	switch statusCode {
	case http.StatusOK:
		resp = common.GeneralError(int32(statusCode), response.Success, message, nil, nil)
	case http.StatusInternalServerError:
		resp = common.GeneralError(int32(statusCode), response.InternalError, message, nil, nil)
	default:
		resp.StatusCode = int32(statusCode)
		resp.StatusMessage = response.ExtendedInfo
		resp.Body = response.CommonError{
			Error: response.ErrorClass{
				Code:    response.ExtendedInfo,
				Message: message,
			},
		}
	}
	return resp
}

// TrackConfigFileChanges monitors the config changes using fsnotfiy
func TrackConfigFileChanges(errChan chan error) {
	eventChan := make(chan interface{})
	format := config.Data.LogFormat
	go common.TrackConfigFileChanges(ConfigFilePath, eventChan, errChan)
	for {
		select {
		case info := <-eventChan:
			l.Log.Info(info) // new data arrives through eventChan channel
			if l.Log.Level != config.Data.LogLevel {
				l.Log.Info("Log level is updated, new log level is ", config.Data.LogLevel)
				l.Log.Logger.SetLevel(config.Data.LogLevel)
			}
			if format != config.Data.LogFormat {
				l.SetFormatter(config.Data.LogFormat)
				format = config.Data.LogFormat
				l.Log.Info("Log format is updated, new log format is ", config.Data.LogFormat)
			}
		case err := <-errChan:
			l.Log.Error(err)
		}
	}
}
