// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var ManifestNotExistErr = errors.New("manifest not existed")

// Manifest represents all basic information of a snapshot.
type Manifest struct {
	Height     uint64 `json:"height"`
	BlockHash  string `json:"hash"`
	FilterId   string `json:"filterId"`
	MerkleRoot string `json:"merkleRoot"`
	Date       string `json:"date"`
	Namespace  string `json:"namespace"`
}

type Manifests []Manifest

/*
	Manifest manipulator
*/

type ManifestRW interface {
	Read(string) (error, Manifest)
	Write(Manifest) error
	List() (error, Manifests)
	Delete(string) error
	Contain(string) bool
	Search(uint64) (error, Manifest)
}

type ManifestHandler struct {
	filePath string
}

func NewManifestHandler(fName string) *ManifestHandler {
	return &ManifestHandler{
		filePath: fName,
	}
}

func (rwc *ManifestHandler) Read(id string) (error, Manifest) {
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return err, Manifest{}
	}
	var manifests Manifests
	if err := json.Unmarshal(buf, &manifests); err != nil {
		return err, Manifest{}
	}
	for _, manifest := range manifests {
		if id == manifest.FilterId {
			return nil, manifest
		}
	}
	return ManifestNotExistErr, Manifest{}
}

func (rwc *ManifestHandler) Write(manifest Manifest) error {
	buf, _ := ioutil.ReadFile(rwc.filePath)
	var manifests Manifests
	if len(buf) != 0 {
		if err := json.Unmarshal(buf, &manifests); err != nil {
			return err
		}
	}
	manifests = append(manifests, manifest)
	if buf, err := json.MarshalIndent(manifests, "", "   "); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(rwc.filePath, buf, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (rwc *ManifestHandler) List() (error, Manifests) {
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return err, nil
	}
	var manifests Manifests
	if err := json.Unmarshal(buf, &manifests); err != nil {
		return err, nil
	}
	return nil, manifests
}

func (rwc *ManifestHandler) Delete(id string) error {
	var deleted bool
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return err
	}
	var manifests Manifests
	if err := json.Unmarshal(buf, &manifests); err != nil {
		return err
	}
	for idx, manifest := range manifests {
		if manifest.FilterId == id {
			manifests = append(manifests[:idx], manifests[idx+1:]...)
			deleted = true
		}
	}
	if deleted {
		if buf, err := json.MarshalIndent(manifests, "", "   "); err != nil {
			return err
		} else {
			if err := ioutil.WriteFile(rwc.filePath, buf, 0644); err != nil {
				return err
			}
			return nil
		}
	} else {
		return ManifestNotExistErr
	}
}

func (rwc *ManifestHandler) Contain(id string) bool {
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return false
	}
	var manifests Manifests
	if err := json.Unmarshal(buf, &manifests); err != nil {
		return false
	}
	for _, manifest := range manifests {
		if manifest.FilterId == id {
			return true
		}
	}
	return false
}

func (rwc *ManifestHandler) Search(height uint64) (error, Manifest) {
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return err, Manifest{}
	}
	var manifests Manifests
	if err := json.Unmarshal(buf, &manifests); err != nil {
		return err, Manifest{}
	}
	for _, manifest := range manifests {
		if manifest.Height == height {
			return nil, manifest
		}
	}
	return ManifestNotExistErr, Manifest{}
}

// ArchiveMeta represents all basic information of historical database.
type ArchiveMeta struct {
	Height       uint64 `json:"height"`
	TransactionN uint64 `json:"transactionNumber"`
	ReceiptN     uint64 `json:"receiptNumber"`
	InvalidTxN   uint64 `json:"invalidTxNumber"`
	LatestUpdate string `json:"latestUpdate"`
}

type ArchiveMetaRW interface {
	Read() (error, ArchiveMeta)
	Write(ArchiveMeta) error
	Exist() bool
}

type ArchiveMetaHandler struct {
	filePath string
}

func NewArchiveMetaHandler(fName string) *ArchiveMetaHandler {
	return &ArchiveMetaHandler{
		filePath: fName,
	}
}

func (rwc *ArchiveMetaHandler) Read() (error, ArchiveMeta) {
	buf, err := ioutil.ReadFile(rwc.filePath)
	if err != nil {
		return err, ArchiveMeta{}
	}
	var meta ArchiveMeta
	if err := json.Unmarshal(buf, &meta); err != nil {
		return err, ArchiveMeta{}
	}
	return nil, meta
}

func (rwc *ArchiveMetaHandler) Write(meta ArchiveMeta) error {
	if buf, err := json.MarshalIndent(meta, "", "   "); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(rwc.filePath, buf, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (rwc *ArchiveMetaHandler) Exist() bool {
	if _, err := os.Stat(rwc.filePath); os.IsNotExist(err) {
		return false
	}
	return true
}
