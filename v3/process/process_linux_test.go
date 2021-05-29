// +build linux

package process

import (
  "context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
  
	"github.com/shirou/gopsutil/v3/internal/common"
	"github.com/stretchr/testify/assert"
)

func Test_Process_splitProcStat(t *testing.T) {
	expectedFieldsNum := 53
	statLineContent := make([]string, expectedFieldsNum-1)
	for i := 0; i < expectedFieldsNum-1; i++ {
		statLineContent[i] = strconv.Itoa(i + 1)
	}

	cases := []string{
		"ok",
		"ok)",
		"(ok",
		"ok )",
		"ok )(",
		"ok )()",
		"() ok )()",
		"() ok (()",
		" ) ok )",
		"(ok) (ok)",
	}

	consideredFields := []int{4, 7, 10, 11, 12, 13, 14, 15, 18, 22, 42}

	commandNameIndex := 2
	for _, expectedName := range cases {
		statLineContent[commandNameIndex-1] = "(" + expectedName + ")"
		statLine := strings.Join(statLineContent, " ")
		t.Run(fmt.Sprintf("name: %s", expectedName), func(t *testing.T) {
			parsedStatLine := splitProcStat([]byte(statLine))
			assert.Equal(t, expectedName, parsedStatLine[commandNameIndex])
			for _, idx := range consideredFields {
				expected := strconv.Itoa(idx)
				parsed := parsedStatLine[idx]
				assert.Equal(
					t, expected, parsed,
					"field %d (index from 1 as in man proc) must be %q but %q is received",
					idx, expected, parsed,
				)
			}
		})
	}
}

func Test_Process_splitProcStat_fromFile(t *testing.T) {
	pid := "68927"
	ppid := "68044"
	statFile := "testdata/linux/proc/" + pid + "/stat"
	contents, err := ioutil.ReadFile(statFile)
	assert.NoError(t, err)
	fields := splitProcStat(contents)
	assert.Equal(t, fields[1], pid)
	assert.Equal(t, fields[2], "test(cmd).sh")
	assert.Equal(t, fields[3], "S")
	assert.Equal(t, fields[4], ppid)
	assert.Equal(t, fields[5], pid)   // pgrp
	assert.Equal(t, fields[6], ppid)  // session
	assert.Equal(t, fields[8], pid)   // tpgrp
	assert.Equal(t, fields[18], "20") // priority
	assert.Equal(t, fields[20], "1")  // num threads
	assert.Equal(t, fields[52], "0")  // exit code
}

func Test_fillFromStatusWithContext(t *testing.T) {
	pids, err := ioutil.ReadDir("testdata/linux/")
	if err != nil {
		t.Error(err)
	}
	f := common.MockEnv("HOST_PROC", "testdata/linux")
	defer f()
	for _, pid := range pids {
		pid, _ := strconv.ParseInt(pid.Name(), 0, 32)
		p, _ := NewProcess(int32(pid))

		if err := p.fillFromStatusWithContext(context.Background()); err != nil {
			t.Error(err)
		}
	}
}