package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillURI(t *testing.T) {

	uri := FillURI("av", "123456", NoteHandler(876))
	fmt.Printf(uri)
	assert.Equal(t, "bilibili://video/123456?cvid=876&locate_note_editing=true", uri)

}
