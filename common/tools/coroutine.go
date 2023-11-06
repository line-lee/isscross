package tool

import (
	"fmt"
	"runtime"
)

func SecureGo(cr func(args ...interface{}), args ...interface{}) {
	go func(arg ...interface{}) {
		defer func() {
			if rc := recover(); rc != nil {
				fmt.Println("安全协程，重启服务", rc)
				for i := 0; i < 20; i++ {
					pc, file, line, ok := runtime.Caller(i)
					if ok {
						f := runtime.FuncForPC(pc)
						fmt.Println(fmt.Sprintf("%s %s %v ", f.Name(), file, line))
					}
				}
			}
		}()
		cr(arg...)
	}(args...)
}
