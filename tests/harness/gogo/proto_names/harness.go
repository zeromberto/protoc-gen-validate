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
		for i:= 0; i < msg.Descriptor().Fields().Len(); i++ {
			fieldNames = append(fieldNames, string(msg.Descriptor().Fields().Get(i).Name()))
		}

		if !containsOneOf(err.Error(), fieldNames) {
			resp(&harness.TestResult{
				Error:  true,
				Reason: fmt.Sprintf("could not find proto field name %v in error message '%s'", fieldNames, err.Error())})
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
