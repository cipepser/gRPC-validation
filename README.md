# gRPC-validation

[Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)でprotobufのvalidationを行う。


## go-proto-validatorのインストール

```sh
❯ go get github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
```

## protobufの定義(validation前)

```proto
syntax = "proto3";

package user;
// import "github.com/mwitkow/go-proto-validators/validator.proto";
service UserService {
  rpc GetUser (Name) returns (User) {}
  rpc GetUsers (Empty) returns (Users) {}
  rpc AddUser (User) returns (Empty) {}
}

message User {
  string name = 1;
  int32 age = 2;
  string phone = 3;
  string mail = 4;
}

message Name {
  string name = 1;
}

message Empty {}

message Users {
  repeated User users = 1;
}
```

## References
* [goのgRPCで便利ツールを使う](https://qiita.com/h3_poteto/items/3a39c41743b4fd87c134)
* [Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)