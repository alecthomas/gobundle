package gobundle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

type Bundled struct {
	Name string
}

var extracted Bundled
 
func tearDown() {
	os.Remove("testbundle.go")
}

func setUp(t *testing.T) {
	cmd := exec.Command(
		"gobundle",
		"--uncompress_on_init",
		"--compress",
		"--package=gobundle",
		"--target=testbundle.go",
		"fixtures/bundled.json",
	)
	_, err := cmd.Output()
	if err != nil {
		t.Fatalf("gobundle command failed: %v\n", err)
	}
}

func TestBundle(t *testing.T) {
	setUp(t)
	bundle, _ := GobundleBundle.Open("fixtures/bundled.json")
	data, _ := ioutil.ReadAll(bundle)
	err := json.Unmarshal(data, &extracted)
	if err != nil {
		t.Fatalf("extraction failed: %v\n", err)
	}
	if extracted.Name != "bundle of joy" {
		t.Fatalf("did not get expected struct, got: %v\n", extracted)
	}
}
