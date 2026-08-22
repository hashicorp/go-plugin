package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	targetPath  = flag.String("target_path", "./", "-target_path=/path/to/save/go/files")
	protoPath   = flag.String("proto_path", "", "-proto_path=/path/to/include/proto/files")
	sourceProto = flag.String("source_proto", "", "-source_proto=/path/to/one/source/proto/file")
)

func checkArgs() {
	if s, err := os.Stat(*targetPath); os.IsNotExist(err) || !s.IsDir() {
		log.Fatalln("target_path not exists(or not a dir): " + *targetPath)
	}
	if s, err := os.Stat(*protoPath); os.IsNotExist(err) || !s.IsDir() {
		log.Fatalln("proto_path not exists(or not a dir): " + *protoPath)
	}
	if s, err := os.Stat(*sourceProto); os.IsNotExist(err) || s.IsDir() {
		log.Fatalln("source_proto not exists(or not a file): " + *sourceProto)
	}
}

func parseFile(protoPath *string, sourceProto *string) *desc.FileDescriptor {
	p := protoparse.Parser{
		IncludeSourceCodeInfo: true,
		ImportPaths:           []string{*protoPath, "."},
	}
	fds, err := p.ParseFiles(*sourceProto)
	if err != nil {
		log.Fatalf("load file [%s] fail, err=%+v", *sourceProto, err)
	}
	if len(fds) != 1 {
		log.Fatalf("load file [%s] fail, count=%d", *sourceProto, len(fds))
	}
	return fds[0]
}

//go:embed template_files/xx.plugin.go.tpl
var templateOfService string

//go:embed template_files/go.mod.tpl
var templateOfGoMod string

//go:embed template_files/callee/go.mod.tpl
var templateOfCalleeGoMod string

//go:embed template_files/callee/callee.go.tpl
var templateOfCalleeDotGo string

//go:embed template_files/callee/plugin/plugin.go.tpl
var templateOfCalleePluginGo string

//go:embed template_files/callee/callee_test.go.tpl
var templateOfCalleeTestGo string

var tService = template.Must(template.New("service").Parse(templateOfService))
var tGoMod = template.Must(template.New("go.mod").Parse(templateOfGoMod))
var tCalleeGoMod = template.Must(template.New("callee_go.mod").Parse(templateOfCalleeGoMod))
var tCalleeDotGo = template.Must(template.New("callee.go").Parse(templateOfCalleeDotGo))
var tCalleePluginGo = template.Must(template.New("callee_plugin.go").Parse(templateOfCalleePluginGo))
var tCalleeTestGo = template.Must(template.New("callee_test.go").Parse(templateOfCalleeTestGo))

func createFile(p string) *os.File {
	out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("create file [%s] fail, err=%s", p, err.Error())
	}
	return out
}

func getMethodOfService(s *desc.ServiceDescriptor) []Method {
	out := make([]Method, 0, len(s.GetMethods()))
	for _, m := range s.GetMethods() {
		out = append(out, Method{
			Name:       makeFirstLetterUpperCase(m.GetName()),
			InputType:  m.GetInputType().GetName(),
			OutputType: m.GetOutputType().GetName(),
		})
	}
	return out
}

func makeFirstLetterUpperCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

var pluginNameRE = regexp.MustCompile(`\d+:["]?([^"]+)["]?`)

func parsePluginName(s string) string {
	arr := pluginNameRE.FindAllStringSubmatch(s, -1)
	if len(arr) == 0 {
		return ""
	}
	if len(arr[0]) < 2 {
		return ""
	}
	return arr[0][1]
}

func getServiceOptionOfPluginName(service *desc.ServiceDescriptor) string {
	opt, ok := service.GetOptions().(*descriptorpb.ServiceOptions)
	if !ok {
		return service.GetName()
	}
	pluginName := parsePluginName(opt.String())
	if len(pluginName) > 0 {
		return pluginName
	}
	// prefix := strconv.Itoa(int(E_PluginName.Field)) + ":"
	// if strings.HasPrefix(opt.String(), prefix) {
	// 	s := opt.String()[len(prefix):]
	// 	if s[0] == '"' {
	// 		s = s[1:]
	// 	}
	// 	if s[len(s)-1] == '"' {
	// 		s = s[:len(s)-1]
	// 	}
	// 	return s
	// }
	return service.GetName()
}

func getServiceInfo(pbFile *desc.FileDescriptor) []Service {
	out := make([]Service, 0, len(pbFile.GetServices()))
	for _, service := range pbFile.GetServices() {
		out = append(out, Service{
			Name:       service.GetName(),
			PluginName: getServiceOptionOfPluginName(service),
			Methods:    getMethodOfService(service),
		})
	}
	return out
}

type Method struct {
	Name       string
	InputType  string
	OutputType string
}

type Service struct {
	Name       string
	PluginName string
	Methods    []Method
}

type pbStruct struct {
	Package         string
	FullPackagePath string
	Services        []Service
}

// var E_PluginName = &proto.ExtensionDesc{
// 	ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
// 	ExtensionType: (*string)(nil),
// 	Field:         51235,
// 	Name:          "grpc_plugin.plugin_name",
// 	Tag:           "bytes,51235,opt,name=plugin_name",
// 	Filename:      "proto/my_test_grpc_plugin.proto",
// }

func genPluginFile(fullPath string, pbFile *desc.FileDescriptor, st *pbStruct) {
	packageName := filepath.Base(fullPath)
	mainFile := createFile(filepath.Join(fullPath, fmt.Sprintf("%s.plugin.go", strings.ToLower(packageName))))
	defer mainFile.Close()
	err := tService.Execute(mainFile, st)
	if err != nil {
		log.Fatalf("write xx.plugin.go fail:%+v", err)
	}
	// go.mod
	goModFile := createFile(filepath.Join(fullPath, "go.mod"))
	defer goModFile.Close()
	err = tGoMod.Execute(goModFile, map[string]string{
		"FullPackagePath": *pbFile.GetFileOptions().GoPackage,
	})
	if err != nil {
		log.Fatalf("write go.mod fail:%+v", err)
	}
}

func mkdirs(d string) {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err = os.MkdirAll(d, os.ModePerm); err != nil {
			log.Fatalf("create dir [%s] fail, err=%s", d, err.Error())
		}
	}
}

func genCalleeFile(fullPath string, pbFile *desc.FileDescriptor, st *pbStruct) {
	calleePath := fullPath + "_callee"
	mkdirs(calleePath)
	// go.mod
	{
		goModFile := createFile(filepath.Join(calleePath, "go.mod"))
		defer goModFile.Close()
		err := tCalleeGoMod.Execute(goModFile, st)
		if err != nil {
			log.Fatalf("write callee go.mod fail:%+v", err)
		}
	}
	// callee.go
	{
		calleeGoFile := createFile(filepath.Join(calleePath, "callee.go"))
		defer calleeGoFile.Close()
		err := tCalleeDotGo.Execute(calleeGoFile, st)
		if err != nil {
			log.Fatalf("write callee/callee.go fail:%+v", err)
		}
	}
	// callee_test.go
	{
		calleeTestFile := createFile(filepath.Join(calleePath, "callee_test.go"))
		defer calleeTestFile.Close()
		err := tCalleeTestGo.Execute(calleeTestFile, st)
		if err != nil {
			log.Fatalf("write callee/callee_test.go fail:%+v", err)
		}
	}
	// plugin/
	pluginPath := calleePath + "/plugin"
	mkdirs(pluginPath)
	{
		calleePluginGoFile := createFile(filepath.Join(pluginPath, "plugin.go"))
		defer calleePluginGoFile.Close()
		err := tCalleePluginGo.Execute(calleePluginGoFile, st)
		if err != nil {
			log.Fatalf("write callee/plugin/plugin.go fail:%+v", err)
		}
	}

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	checkArgs()
	pbFile := parseFile(protoPath, sourceProto)
	// log.Println(*pbFile.GetFileOptions().GoPackage)
	fullPath := filepath.Join(*targetPath, *pbFile.GetFileOptions().GoPackage)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Println("mkdir:", fullPath)
		if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
			log.Fatalln("mkdir fail: err=", err.Error())
		}
	}
	// goFilePath := filepath.Dir(fullPath)
	// log.Println(goFilePath)
	log.Println(fullPath)
	packageName := filepath.Base(fullPath)
	st := &pbStruct{
		Package:         packageName,
		FullPackagePath: *pbFile.GetFileOptions().GoPackage,
		Services:        getServiceInfo(pbFile),
	}
	genPluginFile(fullPath, pbFile, st)
	genCalleeFile(fullPath, pbFile, st)
}
