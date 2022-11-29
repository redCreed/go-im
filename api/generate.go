package api

//go:generate protoc -I. -I ../../      --go_out=plugins=grpc:.  --go_opt=paths=source_relative  connect/connect.proto

// 第一个 -I.表示工作目录   -I ../../表示引入的proto的地址   --go_out=paths=source_relative:logic/ 相对路径，将编译后的文件输出到某个目录
//go:generate protoc -I. -I ../../  --go_out=plugins=grpc:. --go_opt=paths=source_relative logic/logic.proto

//go:generate protoc  -I.  --go_out=plugins=grpc:. --go_opt=paths=source_relative protocol/protocol.proto

//  按道理可以生成，但是确实无法生成
//go:generate protoc -I. -I$GOPATH/src --go_out=plugins=grpc:. --go_opt=paths=source_relative protocol/protocol.proto
//go:generate protoc -I. -I$GOPATH/src --go_out=plugins=grpc:. --go_opt=paths=source_relative comet/comet.proto
//go:generate protoc -I. -I$GOPATH/src --go_out=plugins=grpc:. --go_opt=paths=source_relative logic/logic.proto
