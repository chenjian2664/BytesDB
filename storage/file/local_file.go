/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"BytesDB/core"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
)

// TODO: shall we check storageId? we are writing active file only

// fileStorage FilePerm defines default file permissions (readable by everyone, writable by owner)
type fileStorage struct {
	activeFile *os.File
	oldFiles   []string
	rootPath   string
	schema     string
	tableName  string
	maxSize    int64
	mutex      sync.RWMutex
}

func NewLocalFileStorage(rootPath, schema, table string) (core.Storage, error) {
	dir := path.Join(rootPath, schema, table)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			// TODO: better message
			panic(err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var fileNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// TODO: it is possible the dir contains other type file
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	var activePath string
	if len(fileNames) == 0 {
		// TODO: add util to unified the naming
		activePath = dir + "/" + fmt.Sprintf("%10d.data", 0)
	} else {
		activePath = dir + "/" + fileNames[len(fileNames)-1]
	}
	// note: append mode
	activeFile, err := os.OpenFile(activePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}

	return &fileStorage{
		activeFile: activeFile,
		oldFiles:   fileNames,
		rootPath:   rootPath,
		schema:     schema,
		tableName:  table,
		// 1MB
		maxSize: 1024 * 1024,
		mutex:   sync.RWMutex{}}, nil
}

func (fio *fileStorage) createAndResetActiveFile() {
	fio.mutex.Lock()
	defer fio.mutex.Unlock()

	old := fio.activeFile.Name()
	err := fio.Close()
	if err != nil {
		panic(err)
	}

	fio.oldFiles = append(fio.oldFiles, old)
	// .data
	next, err := strconv.ParseInt(old[:len(old)-5], 10, 64)
	if err != nil {
		panic(err)
	}

	activePath := path.Join(fio.rootPath, fio.schema, fio.tableName, fmt.Sprintf("%10d.data", next))
	// note: append mode
	activeFile, err := os.OpenFile(activePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	fio.activeFile = activeFile
}

func (fio *fileStorage) Read(buf core.Bytes, offset int64) (int, error) {
	return fio.activeFile.ReadAt(buf, offset)
}

func (fio *fileStorage) Write(buf core.Bytes) (int, error) {
	size, err := fio.Size()
	if err != nil {
		panic(err)
	}
	if size+int64(len(buf)) > fio.maxSize {
		err := fio.Flush()
		if err != nil {
			// TODO
			panic(err)
		}
		err = fio.Flush()
		if err != nil {
			panic(err)
		}
		fio.createAndResetActiveFile()
	}

	return fio.activeFile.Write(buf)
}

func (fio *fileStorage) Flush() error {
	return fio.activeFile.Sync()
}

func (fio *fileStorage) Close() error {
	return fio.activeFile.Close()
}

func (fio *fileStorage) Size() (int64, error) {
	stat, err := fio.activeFile.Stat()
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// CurrentStorageId TODO do we need storage expose more info?
func (fio *fileStorage) CurrentStorageId() core.StorageId {
	return core.StorageId{
		Schema: fio.schema,
		Table:  fio.tableName,
	}
}

func (fio *fileStorage) RemoveAll() error {
	path := path.Join(fio.rootPath, fio.schema, fio.tableName)
	return os.RemoveAll(path)
}
