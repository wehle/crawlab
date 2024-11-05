package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var testLogDir string

func setupFileDriverTest() {
	var err error
	testLogDir, err = os.MkdirTemp("", "crawlab-test-logs")
	if err != nil {
		panic(err)
	}
	// Set the log path in viper configuration
	viper.Set("log.path", testLogDir)
}

func cleanupFileDriverTest() {
	_ = os.RemoveAll(testLogDir)
	// Reset the log path in viper configuration
	viper.Set("log.path", "")
}

func TestFileDriver_WriteLine(t *testing.T) {
	setupFileDriverTest()
	t.Cleanup(cleanupFileDriverTest)

	d := newFileLogDriver()
	defer d.Close()

	id := primitive.NewObjectID()

	err := d.WriteLine(id.Hex(), "it works")
	require.Nil(t, err)

	logFilePath := filepath.Join(testLogDir, id.Hex(), "log.txt")
	require.FileExists(t, logFilePath)
	text, err := os.ReadFile(logFilePath)
	require.Nil(t, err)
	require.Equal(t, "it works\n", string(text))
}

func TestFileDriver_WriteLines(t *testing.T) {
	setupFileDriverTest()
	t.Cleanup(cleanupFileDriverTest)

	d := newFileLogDriver()
	defer d.Close()

	id := primitive.NewObjectID()

	for i := 0; i < 100; i++ {
		err := d.WriteLine(id.Hex(), "it works")
		require.Nil(t, err)
	}

	logFilePath := filepath.Join(testLogDir, id.Hex(), "log.txt")
	require.FileExists(t, logFilePath)
	text, err := os.ReadFile(logFilePath)
	require.Nil(t, err)
	require.Contains(t, string(text), "it works\n")
	lines := strings.Split(string(text), "\n")
	require.Equal(t, 101, len(lines))
}

func TestFileDriver_Find(t *testing.T) {
	setupFileDriverTest()
	t.Cleanup(cleanupFileDriverTest)

	d := newFileLogDriver()
	defer d.Close()

	id := primitive.NewObjectID()

	batch := 1000
	var lines []string
	for i := 0; i < 10; i++ {
		for j := 0; j < batch; j++ {
			line := fmt.Sprintf("line: %d", i*batch+j+1)
			lines = append(lines, line)
		}
		err := d.WriteLines(id.Hex(), lines)
		require.Nil(t, err)
		lines = []string{}
	}

	driver := d

	lines, err := driver.Find(id.Hex(), "", 0, 10)
	require.Nil(t, err)
	require.Equal(t, 10, len(lines))
	require.Equal(t, "line: 1", lines[0])
	require.Equal(t, "line: 10", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 0, 1)
	require.Nil(t, err)
	require.Equal(t, 1, len(lines))
	require.Equal(t, "line: 1", lines[0])
	require.Equal(t, "line: 1", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 0, 1000)
	require.Nil(t, err)
	require.Equal(t, 1000, len(lines))
	require.Equal(t, "line: 1", lines[0])
	require.Equal(t, "line: 1000", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 1000, 1000)
	require.Nil(t, err)
	require.Equal(t, 1000, len(lines))
	require.Equal(t, "line: 1001", lines[0])
	require.Equal(t, "line: 2000", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 1001, 1000)
	require.Nil(t, err)
	require.Equal(t, 1000, len(lines))
	require.Equal(t, "line: 1002", lines[0])
	require.Equal(t, "line: 2001", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 1001, 999)
	require.Nil(t, err)
	require.Equal(t, 999, len(lines))
	require.Equal(t, "line: 1002", lines[0])
	require.Equal(t, "line: 2000", lines[len(lines)-1])

	lines, err = driver.Find(id.Hex(), "", 999, 2001)
	require.Nil(t, err)
	require.Equal(t, 2001, len(lines))
	require.Equal(t, "line: 1000", lines[0])
	require.Equal(t, "line: 3000", lines[len(lines)-1])

	cleanupFileDriverTest()
}
