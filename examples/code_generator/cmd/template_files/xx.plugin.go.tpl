package {{.Package}}

import (
	"context"

	"github.com/hashicorp/go-plugin"

	"google.golang.org/grpc"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

{{- range $item := .Services }}

type {{.Name}}Plugin interface{
{{$CurService := .Name}}
{{- range $item := .Methods }}
    {{.Name}}(req *{{.InputType}}) *{{.OutputType}}
{{- end }}
}

type GRPC{{.Name}} struct {
	plugin.Plugin
	Impl {{.Name}}Plugin
}

func (p *GRPC{{.Name}}) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	Register{{.Name}}Server(server, &GPRC{{.Name}}ServerWrapper{impl: p.Impl})
	return nil
}

func (p *GRPC{{.Name}}) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &GRPC{{.Name}}ClientWrapper{client: New{{.Name}}Client(conn)}, nil
}

type GPRC{{.Name}}ServerWrapper struct {
	impl {{.Name}}Plugin
	Unimplemented{{.Name}}Server
}

{{$CurService := .Name}}
{{- range $item := .Methods }}
func (g *GPRC{{$CurService}}ServerWrapper) {{.Name}}(ctx context.Context, req *{{.InputType}}) (*{{.OutputType}}, error) {
	return g.impl.{{.Name}}(req), nil
}
{{- end }}

type GRPC{{.Name}}ClientWrapper struct {
	client {{.Name}}Client
}

{{- range $item := .Methods }}
func (g *GRPC{{$CurService}}ClientWrapper) {{.Name}}(ctx context.Context, req *{{.InputType}}) (*{{.OutputType}}, error) {
	return g.client.{{.Name}}(ctx, req)
}
{{- end }}

{{- end }}
