package open

// 生成 gRPC && bm 代码
//go:generate kratos tool protoc --grpc --bm  open.proto

//go:generate kratos tool protoc --swagger  open.proto
