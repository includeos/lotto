package util

import "testing"

func TestOutputWriter(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		content  [][]string
	}{
		{"standard", "testFile.csv", [][]string{{"1", "2", "3"}, {"a", "b", "c"}}},
		{"extended", "testFile.csv", [][]string{{"1", "2", "3"}, {"next", "line", "please"}}},
		{"extended2", "testFile.csv", [][]string{{"1", "2", "3"}, {"next2", "line", "please"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := outputWriter(tt.content, tt.filePath); err != nil {
				t.Fatal("error writing output", err)
			}
		})
	}

}
