package alias

import (
	"os"
	"testing"

	"go-utils/filesys"
)

func TestAliases(t *testing.T) {
	err := filesys.CopyFile("test1.json", "t.json")
	if err != nil {
		t.Fatalf("Could not copy test1 %v", err)
	}
	defer os.Remove("t.json")
	aliases, err := ReadAliases("t.json")
	if err != nil {
		t.Fatalf("Could not read file %v", err)
	}
	if aliases.Resolve("ml") != "mlx.hpc.uio.no" {
		t.Fatalf("Resolve ml")
	}
	if aliases.Resolve("fox") != "fox.educloud.no" {
		t.Fatalf("Resolve fox")
	}
	if aliases.Resolve("saga") != "saga" {
		t.Fatalf("Resolve saga")
	}

	err = filesys.CopyFile("test2.json", "t.json")
	if err != nil {
		t.Fatalf("Could not copy test2 %v", err)
	}
	err = aliases.Reread()
	if err != nil {
		t.Fatalf("Could not re-read file %v", err)
	}
	if aliases.Resolve("ml") != "mlx.hpc.uio.no" {
		t.Fatalf("Resolve #2 ml")
	}
	if aliases.Resolve("fox") != "fox" {
		t.Fatalf("Resolve #2 fox")
	}
	if aliases.Resolve("saga") != "saga.sigma2.no" {
		t.Fatalf("Resolve #2 saga")
	}
}
