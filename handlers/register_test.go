package handlers

import (
	"github.com/Quaver/api2/config"
	"testing"
)

func TestCheckFlaggedText(t *testing.T) {
	err := config.Load("../config.json")

	if err != nil {
		t.Fatal(err)
	}

	isFlagged, err := isTextFlagged("hello")

	if err != nil {
		t.Fatal(err)
	}

	if isFlagged {
		t.Fatal("Expected text to not be flagged")
	}

	isFlagged, err = isTextFlagged("n*gger")
	
	if err != nil {
		t.Fatal(err)
	}

	if !isFlagged {
		t.Fatal("Expected text to not be flagged")
	}
}
