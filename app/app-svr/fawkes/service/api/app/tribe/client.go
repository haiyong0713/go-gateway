package tribe

// 生成 gRPC && bm 代码
//go:generate kratos tool protoc --grpc --bm  tribe.proto

//go:generate kratos tool protoc --swagger tribe.proto
