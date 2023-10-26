package gen

//go:generate buf generate
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --target gen/ent ./ent/schema
