package plugin

import (
	pb "{{.FullPackagePath}}"
)

{{- range $item := .Services }}

type {{.Name}}Service struct{}

{{$CurService := .Name}}
{{- range $item := .Methods }}
// {{.Name}} Implement the interface of grpc
// don't use error to return any info to caller
func (p *{{$CurService}}Service) {{.Name}}(req *pb.{{.InputType}}) *pb.{{.OutputType}} {
	// todo: add logic here
	return &pb.Response{}
}
{{- end }}
{{- end }}
