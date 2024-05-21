package api
import (
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestZipAndUnzip(t *testing.T) {
	DoZip("../cmd", "../cmd2.zip")
	DoUnzip("../cmd2.zip", "../cmd3")
}