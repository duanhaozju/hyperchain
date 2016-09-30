// author: Lizhong kuang
// date: 2016-09-29
package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// AttributesMetadataEntry is an entry within the metadata that store an attribute name with its respective key.
type AttributesMetadataEntry struct {
	AttributeName string `protobuf:"bytes,1,opt,name=AttributeName" json:"AttributeName,omitempty"`
	AttributeKey  []byte `protobuf:"bytes,2,opt,name=AttributeKey,proto3" json:"AttributeKey,omitempty"`
}

func (m *AttributesMetadataEntry) Reset()         { *m = AttributesMetadataEntry{} }
func (m *AttributesMetadataEntry) String() string { return proto.CompactTextString(m) }
func (*AttributesMetadataEntry) ProtoMessage()    {}

// AttributesMetadata holds both the original metadata bytes and the metadata required to access attributes.
type AttributesMetadata struct {
	// Original metadata bytes
	Metadata []byte `protobuf:"bytes,1,opt,name=Metadata,proto3" json:"Metadata,omitempty"`
	// Entries for each attributes considered.
	Entries []*AttributesMetadataEntry `protobuf:"bytes,2,rep,name=Entries" json:"Entries,omitempty"`
}

func (m *AttributesMetadata) Reset()         { *m = AttributesMetadata{} }
func (m *AttributesMetadata) String() string { return proto.CompactTextString(m) }
func (*AttributesMetadata) ProtoMessage()    {}

func (m *AttributesMetadata) GetEntries() []*AttributesMetadataEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}
