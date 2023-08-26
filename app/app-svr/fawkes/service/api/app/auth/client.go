package auth

// 生成 gRPC && bm 代码
//go:generate kratos tool protoc --grpc --bm  auth.proto

//go:generate kratos tool protoc --swagger  auth.proto
