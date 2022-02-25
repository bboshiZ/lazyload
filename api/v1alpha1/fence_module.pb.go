// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: fence_module.proto

package v1alpha1

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Fence struct {
	// service ports enable lazyload
	WormholePort []string `protobuf:"bytes,1,rep,name=wormholePort,proto3" json:"wormholePort,omitempty"`
	// whether enable ServiceFence auto generating
	// default value is false
	DisableAutoFence     bool     `protobuf:"varint,2,opt,name=disableAutoFence,proto3" json:"disableAutoFence,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Fence) Reset()         { *m = Fence{} }
func (m *Fence) String() string { return proto.CompactTextString(m) }
func (*Fence) ProtoMessage()    {}
func (*Fence) Descriptor() ([]byte, []int) {
	return fileDescriptor_8eebc4b237a55c9b, []int{0}
}
func (m *Fence) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Fence.Unmarshal(m, b)
}
func (m *Fence) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Fence.Marshal(b, m, deterministic)
}
func (m *Fence) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Fence.Merge(m, src)
}
func (m *Fence) XXX_Size() int {
	return xxx_messageInfo_Fence.Size(m)
}
func (m *Fence) XXX_DiscardUnknown() {
	xxx_messageInfo_Fence.DiscardUnknown(m)
}

var xxx_messageInfo_Fence proto.InternalMessageInfo

func (m *Fence) GetWormholePort() []string {
	if m != nil {
		return m.WormholePort
	}
	return nil
}

func (m *Fence) GetDisableAutoFence() bool {
	if m != nil {
		return m.DisableAutoFence
	}
	return false
}

func init() {
	proto.RegisterType((*Fence)(nil), "slime.microservice.lazyload.v1alpha1.Fence")
}

func init() { proto.RegisterFile("fence_module.proto", fileDescriptor_8eebc4b237a55c9b) }

var fileDescriptor_8eebc4b237a55c9b = []byte{
	// 171 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4a, 0x4b, 0xcd, 0x4b,
	0x4e, 0x8d, 0xcf, 0xcd, 0x4f, 0x29, 0xcd, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x52,
	0x29, 0xce, 0xc9, 0xcc, 0x4d, 0xd5, 0xcb, 0xcd, 0x4c, 0x2e, 0xca, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb,
	0x4c, 0x4e, 0xd5, 0xcb, 0x49, 0xac, 0xaa, 0xcc, 0xc9, 0x4f, 0x4c, 0xd1, 0x2b, 0x33, 0x4c, 0xcc,
	0x29, 0xc8, 0x48, 0x34, 0x54, 0x0a, 0xe7, 0x62, 0x75, 0x03, 0xe9, 0x15, 0x52, 0xe2, 0xe2, 0x29,
	0xcf, 0x2f, 0xca, 0xcd, 0xc8, 0xcf, 0x49, 0x0d, 0xc8, 0x2f, 0x2a, 0x91, 0x60, 0x54, 0x60, 0xd6,
	0xe0, 0x0c, 0x42, 0x11, 0x13, 0xd2, 0xe2, 0x12, 0x48, 0xc9, 0x2c, 0x4e, 0x4c, 0xca, 0x49, 0x75,
	0x2c, 0x2d, 0xc9, 0x07, 0xeb, 0x93, 0x60, 0x52, 0x60, 0xd4, 0xe0, 0x08, 0xc2, 0x10, 0x77, 0xd2,
	0x8b, 0xd2, 0x81, 0x38, 0x20, 0x33, 0x5f, 0x1f, 0xcc, 0xd0, 0x87, 0xb8, 0xae, 0x58, 0x1f, 0xe6,
	0x08, 0xfd, 0xc4, 0x82, 0x4c, 0x7d, 0x98, 0x43, 0x92, 0xd8, 0xc0, 0xae, 0x36, 0x06, 0x04, 0x00,
	0x00, 0xff, 0xff, 0xe4, 0xea, 0x54, 0xbd, 0xcb, 0x00, 0x00, 0x00,
}