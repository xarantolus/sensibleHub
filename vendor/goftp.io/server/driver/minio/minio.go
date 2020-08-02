// Copyright 2020 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package minio

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v6"
	"goftp.io/server/core"
)

var (
	_ core.Driver = &Driver{}
)

// Driver implements Driver to store files in minio
type Driver struct {
	client *minio.Client
	perm   core.Perm
	bucket string
}

func buildMinioPath(p string) string {
	return strings.TrimPrefix(p, "/")
}

func buildMinioDir(p string) string {
	v := buildMinioPath(p)
	if !strings.HasSuffix(v, "/") {
		return v + "/"
	}
	return v
}

type myPerm struct {
	core.Perm
	isDir bool
}

func (m *myPerm) GetMode(user string) (os.FileMode, error) {
	mode, err := m.Perm.GetMode(user)
	if err != nil {
		return 0, err
	}
	if m.isDir {
		return mode | os.ModeDir, nil
	}
	return mode, nil
}

type minioFileInfo struct {
	p    string
	info minio.ObjectInfo
	perm core.Perm
}

func (m *minioFileInfo) Name() string {
	return m.p
}

func (m *minioFileInfo) Size() int64 {
	return m.info.Size
}

func (m *minioFileInfo) Mode() os.FileMode {
	mode, _ := m.perm.GetMode(m.p)
	return mode
}

func (m *minioFileInfo) ModTime() time.Time {
	return m.info.LastModified
}

func (m *minioFileInfo) IsDir() bool {
	return m.Mode().IsDir()
}

func (m *minioFileInfo) Sys() interface{} {
	return nil
}

func (m *minioFileInfo) Owner() string {
	owner, _ := m.perm.GetOwner(m.p)
	return owner
}

func (m *minioFileInfo) Group() string {
	group, _ := m.perm.GetGroup(m.p)
	return group
}

func (driver *Driver) isDir(path string) (bool, error) {
	p := buildMinioDir(path)

	info, err := driver.client.StatObject(driver.bucket, p, minio.StatObjectOptions{})
	if err != nil {
		doneCh := make(chan struct{})
		objectCh := driver.client.ListObjects(driver.bucket, p, false, doneCh)
		for object := range objectCh {
			if strings.HasPrefix(object.Key, p) {
				close(doneCh)
				return true, nil
			}
		}

		close(doneCh)
		return false, nil
	}

	return strings.HasSuffix(info.Key, "/"), nil
}

// Stat implements Driver
func (driver *Driver) Stat(path string) (core.FileInfo, error) {
	if path == "/" {
		return &minioFileInfo{
			p:    "/",
			perm: &myPerm{driver.perm, true},
		}, nil
	}

	p := buildMinioPath(path)
	objInfo, err := driver.client.StatObject(driver.bucket, p, minio.StatObjectOptions{})
	if err != nil {
		if isDir, err := driver.isDir(p); err != nil {
			return nil, err
		} else if isDir {
			return &minioFileInfo{
				p:    path,
				perm: &myPerm{driver.perm, true},
			}, nil
		}
		return nil, errors.New("Not a directory")
	}
	isDir := strings.HasSuffix(objInfo.Key, "/")
	return &minioFileInfo{
		p:    p,
		info: objInfo,
		perm: &myPerm{driver.perm, isDir},
	}, nil
}

// ListDir implements Driver
func (driver *Driver) ListDir(path string, callback func(core.FileInfo) error) error {
	doneCh := make(chan struct{})
	defer close(doneCh)

	p := buildMinioDir(path)
	if p == "/" {
		p = ""
	}
	objectCh := driver.client.ListObjects(driver.bucket, p, false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}

		// ignore itself
		if object.Key == p {
			continue
		}

		isDir := strings.HasSuffix(object.Key, "/")
		info := minioFileInfo{
			p:    strings.TrimPrefix(object.Key, p),
			info: object,
			perm: &myPerm{driver.perm, isDir},
		}

		if err := callback(&info); err != nil {
			return err
		}
	}

	return nil
}

// DeleteDir implements Driver
func (driver *Driver) DeleteDir(path string) error {
	doneCh := make(chan struct{})
	defer close(doneCh)

	p := buildMinioPath(path)
	objectCh := driver.client.ListObjects(driver.bucket, p, true, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}

		if err := driver.client.RemoveObject(driver.bucket, object.Key); err != nil {
			return err
		}
	}
	return nil
}

// DeleteFile implements Driver
func (driver *Driver) DeleteFile(path string) error {
	return driver.client.RemoveObject(driver.bucket, buildMinioPath(path))
}

// Rename implements Driver
func (driver *Driver) Rename(fromPath string, toPath string) error {
	src := minio.NewSourceInfo(driver.bucket, buildMinioPath(fromPath), nil)
	dst, err := minio.NewDestinationInfo(driver.bucket, buildMinioPath(toPath), nil, nil)
	if err != nil {
		return err
	}

	if err := driver.client.CopyObject(dst, src); err != nil {
		return err
	}

	return driver.client.RemoveObject(driver.bucket, buildMinioPath(fromPath))
}

// MakeDir implements Driver
func (driver *Driver) MakeDir(path string) error {
	dirPath := buildMinioDir(path)
	_, err := driver.client.PutObject(driver.bucket, dirPath, nil, 0, minio.PutObjectOptions{})
	return err
}

// GetFile implements Driver
func (driver *Driver) GetFile(path string, offset int64) (int64, io.ReadCloser, error) {
	var opts = minio.GetObjectOptions{}
	object, err := driver.client.GetObject(driver.bucket, buildMinioPath(path), opts)
	if err != nil {
		return 0, nil, err
	}
	object.Seek(offset, io.SeekStart)

	info, err := object.Stat()
	if err != nil {
		return 0, nil, err
	}

	return info.Size - offset, object, nil
}

// PutFile implements Driver
func (driver *Driver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {
	p := buildMinioPath(destPath)
	if !appendData {
		return driver.client.PutObject(driver.bucket, p, data, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	}

	tempFile := p + ".tmp"
	//tempDstFile := p + ".dst"
	defer func() {
		if err := driver.DeleteFile(tempFile); err != nil {
			log.Println(err)
		}
		/*if err := driver.DeleteFile(tempDstFile); err != nil {
			log.Println(err)
		}*/
	}()

	size, err := driver.client.PutObject(driver.bucket, tempFile, data, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return size, err
	}

	var srcs = []minio.SourceInfo{
		minio.NewSourceInfo(driver.bucket, tempFile, nil),
		minio.NewSourceInfo(driver.bucket, p, nil),
	}
	dst, err := minio.NewDestinationInfo(driver.bucket, p, nil, nil)
	if err != nil {
		return 0, err
	}

	return size, driver.client.ComposeObject(dst, srcs)
}

// DriverFactory implements DriverFactory
type DriverFactory struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	useSSL          bool
	location        string
	bucket          string
	perm            core.Perm
}

// NewDriverFactory creates a DriverFactory implementation
func NewDriverFactory(endpoint, accessKeyID, secretAccessKey, location, bucket string, useSSL bool, perm core.Perm) *DriverFactory {
	return &DriverFactory{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		useSSL:          useSSL,
		location:        location,
		bucket:          bucket,
		perm:            perm,
	}
}

// NewDriver implements DriverFactory
func (factory *DriverFactory) NewDriver() (core.Driver, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(factory.endpoint, factory.accessKeyID, factory.secretAccessKey, factory.useSSL)
	if err != nil {
		return nil, err
	}

	if err = minioClient.MakeBucket(factory.bucket, factory.location); err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(factory.bucket)
		if !exists || errBucketExists != nil {
			return nil, err
		}
	}

	return &Driver{
		client: minioClient,
		bucket: factory.bucket,
		perm:   factory.perm,
	}, nil
}
