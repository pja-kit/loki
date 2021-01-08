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
	testString := "TestStringÆØÅ"
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	testutils.InitRandom()
	dirName := "/tmp/" + testutils.RandName()
	err := os.MkdirAll(dirName, 0750)
	if err != nil {
		t.Fatal(err)
	}
	positionsFileName := dirName + "/positions.yml"
	defer func() { _ = os.RemoveAll(dirName) }()
	for _, s := range encodings {
		f, err := os.Create(dirName+"/"+s+".log")
		if err != nil {
			t.Fatal("Could not create file", err)
		}
		defer f.Close()
        writeWithEncoding(f,testString,s)
	}
	ps, _ := positions.New(logger, positions.Config{
		SyncPeriod:    10 * time.Second,
		PositionsFile: positionsFileName,
	})
	for _,s := range encodings {
		client := fake.New(func() {})
		defer client.Stop()
		tailer, _ := newTailer(logger, client, ps, dirName+"/"+s+".log",findEncoding(s))
		defer tailer.stop()
		countdown := 10000
		for len(client.Received()) != 1 && countdown > 0 {
			time.Sleep(1 * time.Millisecond)
			countdown--
		}
		// Assert the number of messages the handler received is correct.
		if len(client.Received()) != 1 {
			t.Error("Handler did not receive the correct number of messages, expected 1 received", len(client.Received()))
		}
		// Spot check one of the messages.
		if client.Received()[0].Entry.Line != testString {
			t.Error("Expected first log message to be "+testString+" but was", client.Received()[0].Entry.Line)
		}

	}

}

func writeWithEncoding(f *os.File, text string, encoding string) {
    if(encoding == "ISO-8859-1") {
	  charmap.ISO8859_1.NewEncoder().Writer(f).Write([]byte(text+"\n"))
	}
	if(encoding == "UTF-8") {
		f.WriteString(text+"\n")
	}
}