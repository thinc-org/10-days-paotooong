package gen

//go:generate buf generate
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --target gen/ent ./ent/schema
//go:generate npx @redocly/cli build-docs gen/openapiv2/paotooong.swagger.json --output=static/docs.html
