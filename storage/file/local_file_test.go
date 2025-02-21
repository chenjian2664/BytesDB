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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TODO: add close() test

func TestNewLocalFileManager(t *testing.T) {
	fileName := "/tmp/local-file-manager-test"
	f, err := NewLocalFileManager(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, f)

	t.Cleanup(func() {
		f.Close()
		os.Remove(fileName)
	})
}

func TestFileIO_Write(t *testing.T) {
	fileName := "/tmp/local-file-write-test"
	f, err := NewLocalFileManager(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, f)

	t.Cleanup(func() {
		f.Close()
		os.Remove(fileName)
	})

	n, err := f.Write([]byte(nil))
	assert.Nil(t, err)
	assert.Equal(t, 0, n)

	n, err = f.Write([]byte("hello world"))
	assert.Nil(t, err)
	assert.Equal(t, 11, n)

	n, err = f.Write([]byte("\nhello world"))
	assert.Nil(t, err)
	assert.Equal(t, 12, n)

	n, err = f.Write([]byte("hello world \n"))
	assert.Nil(t, err)
	assert.Equal(t, 13, n)

	n, err = f.Write([]byte("你好"))
	assert.Nil(t, err)
	assert.Equal(t, 6, n)

	n, err = f.Write([]byte("😂"))
	assert.Nil(t, err)
	assert.Equal(t, 4, n)
}

func TestFileIO_Read(t *testing.T) {
	fileName := "/tmp/local-file-read-test"
	f, err := NewLocalFileManager(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, f)

	t.Cleanup(func() {
		f.Close()
		os.Remove(fileName)
	})

	idx := int64(0)

	bs := []byte("hello world")
	n, err := f.Write(bs)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), n)
	buf := make([]byte, len(bs))
	r, err := f.Read(buf, idx)
	assert.Nil(t, err)
	assert.Equal(t, len(buf), r)
	assert.Equal(t, bs, buf)
	idx += int64(r)

	bs = []byte("你好")
	n, err = f.Write(bs)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), n)
	buf = make([]byte, len(bs))
	r, err = f.Read(buf, idx)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), r)
	idx += int64(r)

	bs = []byte("😂")
	n, err = f.Write(bs)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), n)
	buf = make([]byte, len(bs))
	r, err = f.Read(buf, idx)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), r)
}

func TestFileIO_Flush(t *testing.T) {
	fileName := "/tmp/local-file-sync-test"
	f, err := NewLocalFileManager(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, f)

	t.Cleanup(func() {
		f.Close()
		os.Remove(fileName)
	})

	bs := []byte("hello world")
	_, err = f.Write(bs)
	assert.Nil(t, err)
	err = f.Flush()
	assert.Nil(t, err)

	f, err = NewLocalFileManager(fileName)
	assert.Nil(t, err)
	assert.NotNil(t, f)

	buf := make([]byte, len(bs))
	r, err := f.Read(buf, 0)
	assert.Nil(t, err)
	assert.Equal(t, len(bs), r)
	assert.Equal(t, bs, buf)
}
