# gRPC-validation

[Go gRPC Middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)と[Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)でgRPCのvalidationを行う。

## インストール

### Go gRPC Middlewareのインストール

```sh
❯ go get github.com/grpc-ecosystem/go-grpc-middleware
```

### Golang ProtoBuf Validator Compilerのインストール

```sh
❯ go get github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
```

## protobufの定義(validation前)

```proto
syntax = "proto3";

package user;
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

### server(validation前)

```go
package main

import (
	"context"
	"errors"
	"log"
	"net"

	pb "github.com/cipepser/gRPC-validation/user"
	"google.golang.org/grpc"
)

type server struct {
	users map[*pb.User]struct{}
	names map[string]struct{}
}

var (
	empty = new(pb.Empty)
)

const (
	port = ":50051"
)

func (s *server) GetUser(ctx context.Context, in *pb.Name) (*pb.User, error) {
	for u := range s.users {
		if u.Name == in.Name {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *server) GetUsers(ctx context.Context, in *pb.Empty) (*pb.Users, error) {
	out := new(pb.Users)
	for u := range s.users {
		out.Users = append(out.Users, u)
	}
	return out, nil
}

func (s *server) AddUser(ctx context.Context, in *pb.User) (*pb.Empty, error) {
	if _, ok := s.names[in.Name]; ok {
		return empty, errors.New("user already exists")
	}
	s.users[in] = struct{}{}
	s.names[in.Name] = struct{}{}
	return empty, nil
}

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s,
		&server{
			users: map[*pb.User]struct{}{},
			names: map[string]struct{}{},
		},
	)
	s.Serve(l)
}
```

### client(validation前)

```go
package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/cipepser/gRPC-validation/user"
	"google.golang.org/grpc"
)

const (
	address = "localhost"
	port    = ":50051"
)

var (
	empty = new(pb.Empty)
)

type client struct {
}

func main() {
	conn, err := grpc.Dial(address+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn)

	u := pb.User{
		Name:  "Bob",
		Age:   24,
		Phone: "",
		Mail:  "",
	}

	_, err = c.AddUser(context.Background(), &u)
	if err != nil {
		log.Fatalf("failed to add user: %v", err)
	}

	resp, err := c.GetUsers(context.Background(), empty)
	if err != nil {
		log.Fatalf("failed to get users: %v", err)
	}
	log.Printf("users:")
	for _, u := range resp.Users {
		fmt.Println(u)
	}
}
```
## validation

### 仕様

|  フィールド | 制約 |
|  ------ | ------ |
|  name | - |
|  age | 0〜150歳 |
|  phone | 携帯電話の正規表現にマッチ |
|  mail | メールアドレスの正規表現にマッチ |

正規表現の正確さは難易度が高いので今回は、[よく使う正規表現はもうググりたくない！](https://qiita.com/dongri/items/2a0a18e253eb5bf9edba)を利用。エスケープを`\`から`\\`にだけ変えている。

[Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)では、
`[(validator.field) = {msg_exists : true}];`とすることで`required`を実現できるが、[proto3で廃止されたこと](https://developers.google.com/protocol-buffers/docs/proto)からも利用しない。

### protobufの定義(validation後)

```proto
syntax = "proto3";

package user;
import "github.com/mwitkow/go-proto-validators/validator.proto";

service UserService {
  rpc GetUser (Name) returns (User) {}
  rpc GetUsers (Empty) returns (Users) {}
  rpc AddUser (User) returns (Empty) {}
}

message User {
  string name = 1;
  int32 age = 2 [(validator.field) = {int_gt: -1, int_lt: 151}];;
  string phone = 3 [(validator.field) = {regex: "^(070|080|090)-\\d{4}-\\d{4}$"}];
  string mail = 4 [(validator.field) = {regex: "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"}];
}

message Name {
  string name = 1;
}

message Empty {}

message Users {
  repeated User users = 1;
}
```

### コンパイル

```sh
❯ cd user
❯ protoc  \
  --proto_path=${GOPATH}/src \
	--proto_path=. \
	--go_out=plugins=grpc:./ \
	--govalidators_out=./ \
	*.proto
```

### 実装

[Go gRPC Middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)でvalidateさせるために、`server.go`に以下を追記する。

```go
// server.go
s := grpc.NewServer(
  grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
    grpc_validator.StreamServerInterceptor(),
  )),
  grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
    grpc_validator.UnaryServerInterceptor(),
  )),
)
```

### 実行

#### 正常系

```go
u := pb.User{
  Name:  "Alice",
  Age:   20,
  Phone: "090-1111-1111",
  Mail:  "alice@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:20:28 users:
name:"Alice" age:20 phone:"090-1111-1111" mail:"alice@example.com"
```

#### 異常系(age: -1歳)

```go
u := pb.User{
  Name:  "Bob",
  Age:   -1,
  Phone: "090-1111-1111",
  Mail:  "bob@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:22:19 failed to add user: rpc error: code = InvalidArgument desc = invalid field Age: value '-1' must be greater than '-1'
exit status 1
```

#### 異常系(age: 200歳)

```go
u := pb.User{
  Name:  "Bob",
  Age:   200,
  Phone: "090-1111-1111",
  Mail:  "bob@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:22:35 failed to add user: rpc error: code = InvalidArgument desc = invalid field Age: value '200' must be less than '151'
exit status 1
```

#### 異常系(phone: 英字)

```go
u := pb.User{
  Name:  "Bob",
  Age:   20,
  Phone: "09a-1111-11112",
  Mail:  "bob@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:23:40 failed to add user: rpc error: code = InvalidArgument desc = invalid field Phone: value '09a-1111-1111' must be a string conforming to regex "^(070|080|090)-\\d{4}-\\d{4}$"
exit status 1
```

#### 異常系(phone: ハイフンなし)

```go
u := pb.User{
  Name:  "Bob",
  Age:   20,
  Phone: "090111111112",
  Mail:  "bob@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:23:48 failed to add user: rpc error: code = InvalidArgument desc = invalid field Phone: value '09011111111' must be a string conforming to regex "^(070|080|090)-\\d{4}-\\d{4}$"
exit status 1
```

#### 異常系(phone: 桁が多い)

```go
u := pb.User{
  Name:  "Bob",
  Age:   20,
  Phone: "090-1111-11112",
  Mail:  "bob@example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:23:55 failed to add user: rpc error: code = InvalidArgument desc = invalid field Phone: value '090-1111-11112' must be a string conforming to regex "^(070|080|090)-\\d{4}-\\d{4}$"
exit status 1
```

#### 異常系(mail: @なし)

```go
u := pb.User{
  Name:  "Bob",
  Age:   20,
  Phone: "090-1111-1111",
  Mail:  "bob.example.com",
}
```

```sh
❯ go run client/client.go
2018/07/22 14:24:40 failed to add user: rpc error: code = InvalidArgument desc = invalid field Mail: value 'bob.example.com' must be a string conforming to regex "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
exit status 1
```

## References
* [goのgRPCで便利ツールを使う](https://qiita.com/h3_poteto/items/3a39c41743b4fd87c134)
* [Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)
* [Go gRPC Middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
* [よく使う正規表現はもうググりたくない！](https://qiita.com/dongri/items/2a0a18e253eb5bf9edba)
* [Language Guide - Protocol buffers](https://developers.google.com/protocol-buffers/docs/proto)