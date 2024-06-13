// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.2
// source: api.proto

package gen_grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GrpcApiClient is the client API for GrpcApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GrpcApiClient interface {
	// 会话
	SessUserLogin(ctx context.Context, in *SessUserLoginReq, opts ...grpc.CallOption) (*SessUserLoginRes, error)
	SessUserLogout(ctx context.Context, in *SessUserLogoutReq, opts ...grpc.CallOption) (*SessUserLogoutRes, error)
	// 用户管理
	UmRegister(ctx context.Context, in *UmRegisterReq, opts ...grpc.CallOption) (*UmRegisterRes, error)
	UmUnregister(ctx context.Context, in *UmUnregisterReq, opts ...grpc.CallOption) (*UmUnregisterRes, error)
	UmAddFriend(ctx context.Context, in *UmAddFriendReq, opts ...grpc.CallOption) (*UmAddFriendRes, error)
	UmDelFriend(ctx context.Context, in *UmDelFriendReq, opts ...grpc.CallOption) (*UmDelFriendRes, error)
	UmGetFriendList(ctx context.Context, in *UmGetFriendListReq, opts ...grpc.CallOption) (*UmGetFriendListRes, error)
	// 聊天
	ChatGetMsgList(ctx context.Context, in *ChatGetMsgListReq, opts ...grpc.CallOption) (*ChatGetMsgListRes, error)
	ChatGetBoxMsgHist(ctx context.Context, in *ChatGetBoxMsgHistReq, opts ...grpc.CallOption) (*ChatGetBoxMsgHistRes, error)
	ChatSendMsg(ctx context.Context, in *ChatSendMsgReq, opts ...grpc.CallOption) (*ChatSendMsgRes, error)
	ChatCreateGroupConv(ctx context.Context, in *ChatCreateGroupConvReq, opts ...grpc.CallOption) (*ChatCreateGroupConvRes, error)
	ChatGetGroupConvList(ctx context.Context, in *ChatGetGroupConvListReq, opts ...grpc.CallOption) (*ChatGetGroupConvListRes, error)
}

type grpcApiClient struct {
	cc grpc.ClientConnInterface
}

func NewGrpcApiClient(cc grpc.ClientConnInterface) GrpcApiClient {
	return &grpcApiClient{cc}
}

func (c *grpcApiClient) SessUserLogin(ctx context.Context, in *SessUserLoginReq, opts ...grpc.CallOption) (*SessUserLoginRes, error) {
	out := new(SessUserLoginRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/SessUserLogin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) SessUserLogout(ctx context.Context, in *SessUserLogoutReq, opts ...grpc.CallOption) (*SessUserLogoutRes, error) {
	out := new(SessUserLogoutRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/SessUserLogout", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) UmRegister(ctx context.Context, in *UmRegisterReq, opts ...grpc.CallOption) (*UmRegisterRes, error) {
	out := new(UmRegisterRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/UmRegister", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) UmUnregister(ctx context.Context, in *UmUnregisterReq, opts ...grpc.CallOption) (*UmUnregisterRes, error) {
	out := new(UmUnregisterRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/UmUnregister", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) UmAddFriend(ctx context.Context, in *UmAddFriendReq, opts ...grpc.CallOption) (*UmAddFriendRes, error) {
	out := new(UmAddFriendRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/UmAddFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) UmDelFriend(ctx context.Context, in *UmDelFriendReq, opts ...grpc.CallOption) (*UmDelFriendRes, error) {
	out := new(UmDelFriendRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/UmDelFriend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) UmGetFriendList(ctx context.Context, in *UmGetFriendListReq, opts ...grpc.CallOption) (*UmGetFriendListRes, error) {
	out := new(UmGetFriendListRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/UmGetFriendList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) ChatGetMsgList(ctx context.Context, in *ChatGetMsgListReq, opts ...grpc.CallOption) (*ChatGetMsgListRes, error) {
	out := new(ChatGetMsgListRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/ChatGetMsgList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) ChatGetBoxMsgHist(ctx context.Context, in *ChatGetBoxMsgHistReq, opts ...grpc.CallOption) (*ChatGetBoxMsgHistRes, error) {
	out := new(ChatGetBoxMsgHistRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/ChatGetBoxMsgHist", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) ChatSendMsg(ctx context.Context, in *ChatSendMsgReq, opts ...grpc.CallOption) (*ChatSendMsgRes, error) {
	out := new(ChatSendMsgRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/ChatSendMsg", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) ChatCreateGroupConv(ctx context.Context, in *ChatCreateGroupConvReq, opts ...grpc.CallOption) (*ChatCreateGroupConvRes, error) {
	out := new(ChatCreateGroupConvRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/ChatCreateGroupConv", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *grpcApiClient) ChatGetGroupConvList(ctx context.Context, in *ChatGetGroupConvListReq, opts ...grpc.CallOption) (*ChatGetGroupConvListRes, error) {
	out := new(ChatGetGroupConvListRes)
	err := c.cc.Invoke(ctx, "/gen_grpc.GrpcApi/ChatGetGroupConvList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GrpcApiServer is the server API for GrpcApi service.
// All implementations must embed UnimplementedGrpcApiServer
// for forward compatibility
type GrpcApiServer interface {
	// 会话
	SessUserLogin(context.Context, *SessUserLoginReq) (*SessUserLoginRes, error)
	SessUserLogout(context.Context, *SessUserLogoutReq) (*SessUserLogoutRes, error)
	// 用户管理
	UmRegister(context.Context, *UmRegisterReq) (*UmRegisterRes, error)
	UmUnregister(context.Context, *UmUnregisterReq) (*UmUnregisterRes, error)
	UmAddFriend(context.Context, *UmAddFriendReq) (*UmAddFriendRes, error)
	UmDelFriend(context.Context, *UmDelFriendReq) (*UmDelFriendRes, error)
	UmGetFriendList(context.Context, *UmGetFriendListReq) (*UmGetFriendListRes, error)
	// 聊天
	ChatGetMsgList(context.Context, *ChatGetMsgListReq) (*ChatGetMsgListRes, error)
	ChatGetBoxMsgHist(context.Context, *ChatGetBoxMsgHistReq) (*ChatGetBoxMsgHistRes, error)
	ChatSendMsg(context.Context, *ChatSendMsgReq) (*ChatSendMsgRes, error)
	ChatCreateGroupConv(context.Context, *ChatCreateGroupConvReq) (*ChatCreateGroupConvRes, error)
	ChatGetGroupConvList(context.Context, *ChatGetGroupConvListReq) (*ChatGetGroupConvListRes, error)
	mustEmbedUnimplementedGrpcApiServer()
}

// UnimplementedGrpcApiServer must be embedded to have forward compatible implementations.
type UnimplementedGrpcApiServer struct {
}

func (UnimplementedGrpcApiServer) SessUserLogin(context.Context, *SessUserLoginReq) (*SessUserLoginRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SessUserLogin not implemented")
}
func (UnimplementedGrpcApiServer) SessUserLogout(context.Context, *SessUserLogoutReq) (*SessUserLogoutRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SessUserLogout not implemented")
}
func (UnimplementedGrpcApiServer) UmRegister(context.Context, *UmRegisterReq) (*UmRegisterRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmRegister not implemented")
}
func (UnimplementedGrpcApiServer) UmUnregister(context.Context, *UmUnregisterReq) (*UmUnregisterRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmUnregister not implemented")
}
func (UnimplementedGrpcApiServer) UmAddFriend(context.Context, *UmAddFriendReq) (*UmAddFriendRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmAddFriend not implemented")
}
func (UnimplementedGrpcApiServer) UmDelFriend(context.Context, *UmDelFriendReq) (*UmDelFriendRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmDelFriend not implemented")
}
func (UnimplementedGrpcApiServer) UmGetFriendList(context.Context, *UmGetFriendListReq) (*UmGetFriendListRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UmGetFriendList not implemented")
}
func (UnimplementedGrpcApiServer) ChatGetMsgList(context.Context, *ChatGetMsgListReq) (*ChatGetMsgListRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetMsgList not implemented")
}
func (UnimplementedGrpcApiServer) ChatGetBoxMsgHist(context.Context, *ChatGetBoxMsgHistReq) (*ChatGetBoxMsgHistRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetBoxMsgHist not implemented")
}
func (UnimplementedGrpcApiServer) ChatSendMsg(context.Context, *ChatSendMsgReq) (*ChatSendMsgRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatSendMsg not implemented")
}
func (UnimplementedGrpcApiServer) ChatCreateGroupConv(context.Context, *ChatCreateGroupConvReq) (*ChatCreateGroupConvRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatCreateGroupConv not implemented")
}
func (UnimplementedGrpcApiServer) ChatGetGroupConvList(context.Context, *ChatGetGroupConvListReq) (*ChatGetGroupConvListRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChatGetGroupConvList not implemented")
}
func (UnimplementedGrpcApiServer) mustEmbedUnimplementedGrpcApiServer() {}

// UnsafeGrpcApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GrpcApiServer will
// result in compilation errors.
type UnsafeGrpcApiServer interface {
	mustEmbedUnimplementedGrpcApiServer()
}

func RegisterGrpcApiServer(s grpc.ServiceRegistrar, srv GrpcApiServer) {
	s.RegisterService(&GrpcApi_ServiceDesc, srv)
}

func _GrpcApi_SessUserLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessUserLoginReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).SessUserLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/SessUserLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).SessUserLogin(ctx, req.(*SessUserLoginReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_SessUserLogout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessUserLogoutReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).SessUserLogout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/SessUserLogout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).SessUserLogout(ctx, req.(*SessUserLogoutReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_UmRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UmRegisterReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).UmRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/UmRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).UmRegister(ctx, req.(*UmRegisterReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_UmUnregister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UmUnregisterReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).UmUnregister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/UmUnregister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).UmUnregister(ctx, req.(*UmUnregisterReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_UmAddFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UmAddFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).UmAddFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/UmAddFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).UmAddFriend(ctx, req.(*UmAddFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_UmDelFriend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UmDelFriendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).UmDelFriend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/UmDelFriend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).UmDelFriend(ctx, req.(*UmDelFriendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_UmGetFriendList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UmGetFriendListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).UmGetFriendList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/UmGetFriendList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).UmGetFriendList(ctx, req.(*UmGetFriendListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_ChatGetMsgList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatGetMsgListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).ChatGetMsgList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/ChatGetMsgList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).ChatGetMsgList(ctx, req.(*ChatGetMsgListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_ChatGetBoxMsgHist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatGetBoxMsgHistReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).ChatGetBoxMsgHist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/ChatGetBoxMsgHist",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).ChatGetBoxMsgHist(ctx, req.(*ChatGetBoxMsgHistReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_ChatSendMsg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatSendMsgReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).ChatSendMsg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/ChatSendMsg",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).ChatSendMsg(ctx, req.(*ChatSendMsgReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_ChatCreateGroupConv_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatCreateGroupConvReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).ChatCreateGroupConv(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/ChatCreateGroupConv",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).ChatCreateGroupConv(ctx, req.(*ChatCreateGroupConvReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _GrpcApi_ChatGetGroupConvList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChatGetGroupConvListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GrpcApiServer).ChatGetGroupConvList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gen_grpc.GrpcApi/ChatGetGroupConvList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GrpcApiServer).ChatGetGroupConvList(ctx, req.(*ChatGetGroupConvListReq))
	}
	return interceptor(ctx, in, info, handler)
}

// GrpcApi_ServiceDesc is the grpc.ServiceDesc for GrpcApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GrpcApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gen_grpc.GrpcApi",
	HandlerType: (*GrpcApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SessUserLogin",
			Handler:    _GrpcApi_SessUserLogin_Handler,
		},
		{
			MethodName: "SessUserLogout",
			Handler:    _GrpcApi_SessUserLogout_Handler,
		},
		{
			MethodName: "UmRegister",
			Handler:    _GrpcApi_UmRegister_Handler,
		},
		{
			MethodName: "UmUnregister",
			Handler:    _GrpcApi_UmUnregister_Handler,
		},
		{
			MethodName: "UmAddFriend",
			Handler:    _GrpcApi_UmAddFriend_Handler,
		},
		{
			MethodName: "UmDelFriend",
			Handler:    _GrpcApi_UmDelFriend_Handler,
		},
		{
			MethodName: "UmGetFriendList",
			Handler:    _GrpcApi_UmGetFriendList_Handler,
		},
		{
			MethodName: "ChatGetMsgList",
			Handler:    _GrpcApi_ChatGetMsgList_Handler,
		},
		{
			MethodName: "ChatGetBoxMsgHist",
			Handler:    _GrpcApi_ChatGetBoxMsgHist_Handler,
		},
		{
			MethodName: "ChatSendMsg",
			Handler:    _GrpcApi_ChatSendMsg_Handler,
		},
		{
			MethodName: "ChatCreateGroupConv",
			Handler:    _GrpcApi_ChatCreateGroupConv_Handler,
		},
		{
			MethodName: "ChatGetGroupConvList",
			Handler:    _GrpcApi_ChatGetGroupConvList_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
