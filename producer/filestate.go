// Copyright 2015-2017 trivago GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package producer

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/trivago/gollum/core"
	"io"
	"strconv"
	"strings"
	"time"
)

// FileState is a helper struct for io.Writer compatible classes that use batch directly for resources
type FileState struct {
	writer       FileStateWriter
	batch        core.MessageBatch
	assembly     core.WriterAssembly
	fileCreated  time.Time
	flushTimeout time.Duration
	logger       logrus.FieldLogger
}

// FileStateWriter is an interface for different file writer like disk, s3, etc.
type FileStateWriter interface {
	io.WriteCloser
	Name() string // base name of the file
	Size() int64  // length in bytes for regular files; system-dependent for others
	//Created() time.Time
	IsAccessible() bool
}

// NewFileState returns a new FileState instance
func NewFileState(maxMessageCount int, modulator core.Modulator, tryFallback func(*core.Message),
	timeout time.Duration, logger logrus.FieldLogger) *FileState {
	return &FileState{
		batch:        core.NewMessageBatch(maxMessageCount),
		flushTimeout: timeout,
		assembly:     core.NewWriterAssembly(nil, tryFallback, modulator),
		logger:       logger,
	}
}

// Flush flush the batch
func (state *FileState) Flush() {
	if state.writer != nil {
		state.assembly.SetWriter(state.writer)
		state.batch.Flush(state.assembly.Write)
	} else {
		state.batch.Flush(state.assembly.Flush)
	}
}

// Close closes batch and writer
func (state *FileState) Close() {
	if state.writer != nil {
		state.assembly.SetWriter(state.writer)
		state.batch.Close(state.assembly.Write, state.flushTimeout)
	} else {
		state.batch.Close(state.assembly.Flush, state.flushTimeout)
	}
	state.writer.Close()
}

// NeedsRotate evaluate if the FileState need to rotate by the FileRotateConfig
func (state *FileState) NeedsRotate(rotate FileRotateConfig, forceRotate bool) (bool, error) {
	// File does not exist?
	if state.writer == nil {
		return true, nil
	}

	// File can be accessed?
	if state.writer.IsAccessible() == false {
		return false, errors.New("Can' access file to rotate")
	}

	// File needs rotation?
	if !rotate.enabled {
		return false, nil
	}

	if forceRotate {
		return true, nil
	}

	// File is too large?
	if state.writer.Size() >= rotate.sizeByte {
		return true, nil // ### return, too large ###
	}

	// File is too old?
	if time.Since(state.fileCreated) >= rotate.timeout {
		return true, nil // ### return, too old ###
	}

	// RotateAt crossed?
	if rotate.atHour > -1 && rotate.atMinute > -1 {
		now := time.Now()
		rotateAt := time.Date(now.Year(), now.Month(), now.Day(), rotate.atHour, rotate.atMinute, 0, 0, now.Location())

		if state.fileCreated.Sub(rotateAt).Minutes() < 0 {
			return true, nil // ### return, too old ###
		}
	}

	// nope, everything is ok
	return false, nil
}

// -- FileRotateConfig --

// FileRotateConfig defines the rotation settings
type FileRotateConfig struct {
	timeout   time.Duration `config:"Rotation/TimeoutMin" default:"1440" metric:"min"`
	sizeByte  int64         `config:"Rotation/SizeMB" default:"1024" metric:"mb"`
	zeroPad   int           `config:"Rotation/ZeroPadding" default:"0"`
	enabled   bool          `config:"Rotation/Enable" default:"false"`
	compress  bool          `config:"Rotation/Compress" default:"false"`
	timestamp string        `config:"Rotation/Timestamp" default:"2006-01-02_15"`
	atHour    int           `config:"Rotation/AtHour" default:"-1"`
	atMinute  int           `config:"Rotation/AtMin" default:"-1"`
}

// Configure method for interface implementation
func (rotate *FileRotateConfig) Configure(conf core.PluginConfigReader) {
	rotateAt := conf.GetString("Rotation/At", "")
	if rotateAt != "" {
		parts := strings.Split(rotateAt, ":")
		rotateAtHour, _ := strconv.ParseInt(parts[0], 10, 8)
		rotateAtMin, _ := strconv.ParseInt(parts[1], 10, 8)

		rotate.atHour = int(rotateAtHour)
		rotate.atMinute = int(rotateAtMin)
	}
}
