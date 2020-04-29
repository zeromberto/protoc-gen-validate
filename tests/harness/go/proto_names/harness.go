package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	_ "github.com/envoyproxy/protoc-gen-validate/tests/harness/cases/go"
	_ "github.com/envoyproxy/protoc-gen-validate/tests/harness/cases/other_package/go"
	harness "github.com/envoyproxy/protoc-gen-validate/tests/harness/go"
)

func main() {
	b, err := ioutil.ReadAll(os.Stdin)
	checkErr(err)

	tc := new(harness.TestCase)
	checkErr(proto.Unmarshal(b, tc))

	da := new(ptypes.DynamicAny)
	checkErr(ptypes.UnmarshalAny(tc.Message, da))
	v := da.Message.(interface {
		Validate() error
	})

	checkMsg(da.Message, v.Validate())
}

func checkMsg(message proto.Message, err error) {
	if err != nil {
		msg := proto.MessageReflect(message)
		var fieldNames []string
		fields := msg.Descriptor().Fields()
		for i := 0; i < fields.Len(); i++ {
			msgPrefix := string(msg.Descriptor().Name()) + "."
			oneof := fields.Get(i).ContainingOneof()
			if oneof == nil {
				fieldNames = append(fieldNames, msgPrefix+string(fields.Get(i).Name()))
			} else {
				fieldNames = append(fieldNames, msgPrefix+string(oneof.Name()))
				for j := 0; j < oneof.Fields().Len(); j++ {
					fieldNames = append(fieldNames, msgPrefix+string(oneof.Fields().Get(i).Name()))
				}
			}
		}

		if !containsOneOf(err.Error(), fieldNames) {
			resp(&harness.TestResult{
				Error:  true,
				Reason: fmt.Sprintf("could not find any of the proto field names %v in error message '%s'", fieldNames, err.Error())})
		} else {
			resp(&harness.TestResult{Reason: err.Error()})
		}
	}

	resp(&harness.TestResult{Valid: true})
}

func containsOneOf(s string, names []string) bool {
	for _, name := range names {
		if strings.Contains(s, name) {
			return true
		}
	}

	return false
}

func checkErr(err error) {
	if err == nil {
		return
	}

	resp(&harness.TestResult{
		Error:  true,
		Reason: err.Error(),
	})
}

func resp(result *harness.TestResult) {
	if b, err := proto.Marshal(result); err != nil {
		log.Fatalf("could not marshal response: %v", err)
	} else if _, err = os.Stdout.Write(b); err != nil {
		log.Fatalf("could not write response: %v", err)
	}

	os.Exit(0)
}
