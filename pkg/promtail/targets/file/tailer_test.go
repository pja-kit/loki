package file

import (
	"github.com/go-kit/kit/log"
	"github.com/grafana/loki/pkg/promtail/client/fake"
	"github.com/grafana/loki/pkg/promtail/positions"
	"github.com/grafana/loki/pkg/promtail/targets/testutils"
	"golang.org/x/text/encoding/charmap"
	"os"
	"testing"
	"time"
)

func TestEncodings(t *testing.T) {
	encodings := []string{"UTF-8","ISO-8859-1"}
	testString := "TestStringÆØÅ\n"
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	testutils.InitRandom()
	dirName := "/tmp/" + testutils.RandName()
	positionsFileName := dirName + "/positions.yml"
	defer func() { _ = os.RemoveAll(dirName) }()
	for _, s := range encodings {
		f, _ := os.Create(dirName+"/"+s+".log")
		defer f.Close()
        writeWithEncoding(f,testString,s)
	}
	ps, _ := positions.New(logger, positions.Config{
		SyncPeriod:    10 * time.Second,
		PositionsFile: positionsFileName,
	})
	client := fake.New(func() {})
	defer client.Stop()
	for _,s := range encodings {
		tailer, _ := newTailer(logger, client, ps, dirName+"/"+s+".log")
		defer tailer.stop()
		// Assert the number of messages the handler received is correct.
		if len(client.Received()) != 1 {
			t.Error("Handler did not receive the correct number of messages, expected 1 received", len(client.Received()))
		}
		// Spot check one of the messages.
		if client.Received()[0].Line != testString {
			t.Error("Expected first log message to be "+testString+" but was", client.Received()[0])
		}

	}

}

func writeWithEncoding(f *os.File, text string, encoding string) {
    if(encoding == "ISO-8859-1") {
	  charmap.ISO8859_1.NewEncoder().Writer(f).Write([]byte(text))
	}
	if(encoding == "UTF-8") {
		f.WriteString(text)
	}
}