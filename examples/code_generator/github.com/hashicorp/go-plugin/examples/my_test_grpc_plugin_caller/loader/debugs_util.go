package loader

import (
	"fmt"
	"runtime"
	"strings"
)

// SourceCodeLoc 获取源码行
// example: SourceCodeLoc(1) 获取当前代码行的行号
func SourceCodeLoc(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return ""
	}
	file = strings.ReplaceAll(file, "\\", "/")
	paths := strings.Split(file, "/")
	if len(paths) > 3 {
		file = strings.Join(paths[len(paths)-3:], "/")
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// WrapError 在错误信息的基础上加上当前行号
func WrapError(err error, infos ...string) error {
	return fmt.Errorf("[%s] %+v, err=%+v", SourceCodeLoc(2), infos, err)
}
