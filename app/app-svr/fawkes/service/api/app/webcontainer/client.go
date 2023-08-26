package webcontainer

// 生成 gRPC && bm 代码
//go:generate kratos tool protoc --grpc --bm  whitelist.proto

//go:generate kratos tool protoc --swagger whitelist.proto
