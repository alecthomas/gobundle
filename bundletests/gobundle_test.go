package bundletests

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/alecthomas/gobundle"
)

type Bundled struct {
	Name string
}

var (
	extracted Bundled
)

func tearDown() {
	os.Remove("testbundle.go")
}

func setUp(t *testing.T) {
	cmd := exec.Command(
		"gobundle",
		"--uncompress_on_init",
		"--compress",
		"--package=bundletests",
		"--target=testbundle.go",
		"fixtures/bundled.json",
	)
	_, err := cmd.Output()
	if err != nil {
		t.Fatalf("gobundle command failed: %v\n", err)
	}
}

func TestBundle(t *testing.T) {
	file, _ := os.Open("fixtures/bundled.json")
	data, _ := ioutil.ReadAll(file)
	BundletestsBundle := gobundle.NewBuilder("bundletests").Add(
		"fixtures/bundled.json", data,
	).Build()
	bundle, _ := BundletestsBundle.Open("fixtures/bundled.json")
	data, _ = ioutil.ReadAll(bundle)
	err := json.Unmarshal(data, &extracted)
	if err != nil {
		t.Fatalf("extraction failed: %v\n", err)
	}
	if extracted.Name != "bundle of joy" {
		t.Fatalf("did not get expected struct, got: %v\n", extracted)
	}
}
