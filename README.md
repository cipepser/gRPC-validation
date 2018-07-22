# gRPC-validation

[Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)でprotobufのvalidationを行う。

## インストール

### go-grpc-middlewareのインストール

```sh
❯ go get github.com/grpc-ecosystem/go-grpc-middleware
```

### go-proto-validatorのインストール

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
|  name | required |
|  age | 0〜150歳 |
|  phone | 携帯電話の正規表現にマッチ |
|  mail | メールアドレスの正規表現にマッチ |

正規表現の正確さは難易度が高いので今回は、[よく使う正規表現はもうググりたくない！](https://qiita.com/dongri/items/2a0a18e253eb5bf9edba)を利用。

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

## References
* [goのgRPCで便利ツールを使う](https://qiita.com/h3_poteto/items/3a39c41743b4fd87c134)
* [Golang ProtoBuf Validator Compiler](https://github.com/mwitkow/go-proto-validators)
* [よく使う正規表現はもうググりたくない！](https://qiita.com/dongri/items/2a0a18e253eb5bf9edba)