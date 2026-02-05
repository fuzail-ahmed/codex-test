param(
  [string]$ProtoDir = "proto/todo/v1"
)

$protoFiles = Get-ChildItem -Path $ProtoDir -Filter "*.proto" | ForEach-Object { $_.FullName }

protoc --proto_path=$ProtoDir `
  --go_out=shared/gen/todo/v1 --go-grpc_out=shared/gen/todo/v1 `
  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative `
  $protoFiles