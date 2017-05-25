package common

import (
	"errors"
	"io/ioutil"
	"encoding/json"
	"os"
)

var ManifestNotExistErr   = errors.New("manifest not existed")

type Manifest struct {
	Height     uint64    `json:"height"`
	FilterId   string    `json:"filterId"`
	MerkleRoot string    `json:"merkleRoot"`
	Date       string    `json:"date"`
	Namespace  string    `json:"namespace"`
}

type Manifests []Manifest
/*
	Manifest manipulator
 */

type ManifestRWC interface {
	Read(string) (error, Manifest)
	Write(Manifest) error
	List() (error, Manifests)
	Delete(string) error
	Contain(string) bool
}

type ManifestHandler struct {
	filePath    string
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

/*
	Archive Meta
 */
type ArchiveMeta struct {
	Height          uint64     `json:"height"`
	TransactionN    uint64     `json:"transactionNumber"`
	ReceiptN        uint64     `json:"receiptNumber"`
	InvalidTxN      uint64     `json:"invalidTxNumber"`
	LatestUpdate    string     `json:"latestUpdate"`
}

type ArchiveMetaRWC interface {
	Read() (error, ArchiveMeta)
	Write(ArchiveMeta) error
	Exist() bool
}

type ArchiveMetaHandler struct {
	filePath    string
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
