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

package storage

import (
	"BytesDB/config"
	"BytesDB/core"
	"BytesDB/storage/file"
	"sync"
)

type StorageType = byte

const (
	Local_File StorageType = iota
)

type StorageManager struct {
	storages map[core.Session]core.Storage
	mutex    sync.RWMutex
	options  *StorageOptions
	typ      StorageType
}

func NewStorageManager(cfg *config.DBConfig) *StorageManager {
	return &StorageManager{
		make(map[core.Session]core.Storage),
		sync.RWMutex{},
		FromDbOptions(cfg),
		resolveStorageType(cfg.StorageType),
	}
}

func resolveStorageType(typ string) StorageType {
	// by default
	if typ == "" {
		return Local_File
	}
	switch typ {
	case "local_file":
		return Local_File
	default:
		panic("unknown storage type")
	}
}

func (sm *StorageManager) Read(session core.Session, position *core.RecordPosition) *core.Record {
	storage := sm.resolveStorage(session)

	// TODO: consider shall we reader header separately, instead of read whole record size
	bytes := make(core.Bytes, position.Size)
	_, err := storage.Read(bytes, position.Position)
	if err != nil {
		panic(err)
	}

	return core.BytesToRecord(bytes)
}

// append
func (sm *StorageManager) Write(session core.Session, record *core.Record) *core.RecordPosition {
	storage := sm.resolveStorage(session)
	bytes := record.Pack()
	write, err := storage.Write(bytes)
	if err != nil {
		panic(err)
	}

	// TODO: maybe we could record the stats here instead of ask storage everytime
	sz, _ := storage.Size()

	return &core.RecordPosition{
		Position: sz - int64(write),
		Size:     write,
	}
}

func (sm *StorageManager) Delete(session core.Session, key core.Bytes) *core.RecordPosition {
	record := &core.Record{
		Key:   key,
		Value: core.Bytes{},
		Type:  core.Deleted,
	}
	return sm.Write(session, record)
}

func (sm *StorageManager) RemoveAllData(sid core.Session) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	_ = sm.resolveStorage(sid).RemoveAll()
}

func (sm *StorageManager) Close() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	for _, storage := range sm.storages {
		_ = storage.Close()
	}
}

func (sm *StorageManager) Size(session core.Session) (int64, error) {
	storage := sm.resolveStorage(session)
	return storage.Size()
}

func (sm *StorageManager) resolveStorage(sid core.Session) core.Storage {
	if _, ok := sm.storages[sid]; !ok {
		sm.initializeStorage(sm.typ, sid)
	}

	storage := sm.storages[sid]
	return storage
}

func (sm *StorageManager) initializeStorage(storageType StorageType, session core.Session) {
	if _, ok := sm.storages[session]; ok {
		return
	}
	switch storageType {
	case Local_File:
		storage, err := file.NewLocalFileStorage(sm.options.rootPath, session.Schema, session.Table)
		if err != nil {
			panic(err)
		}
		sm.storages[session] = storage
	default:
		// TODO: create error
		panic("storage type not supported: " + string(storageType))
	}
}
