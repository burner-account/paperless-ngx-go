package paperless

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DocumentFileTypeCount struct {
	MIMEType      string `json:"mime_type"`
	MIMETypeCount int    `json:"mime_type_count"`
}

func (d DocumentFileTypeCount) String() string {
	return fmt.Sprintf(`
	  - mime type: %s
	    mime type count: %d`,
		d.MIMEType,
		d.MIMETypeCount,
	)
}

type Statistics struct {
	DocumentsTotal        int                     `json:"documents_total"`
	DocumentsInbox        int                     `json:"documents_inbox"`
	InboxTag              string                  `json:"inbox_tag"`
	InboxTags             []string                `json:"inbox_tags"`
	DocumentFileTypeCount []DocumentFileTypeCount `json:"document_file_type_counts"`
	CharacterCount        int                     `json:"character_count"`
	TagCount              int                     `json:"tag_count"`
	CorrespondentCount    int                     `json:"correspondent_count"`
	DocumentTypeCount     int                     `json:"document_type_count"`
	StoragePathCount      int                     `json:"storage_path_count"`
	CurrentASN            int                     `json:"current_asn"`
}

func NewStatistics(raw map[string]interface{}) (*Statistics, error) {
	bytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	output := &Statistics{}
	err = json.Unmarshal(bytes, output)
	return output, err
}

func (s Statistics) String() string {
	dftCounts := make([]string, 0)
	for _, dft := range s.DocumentFileTypeCount {
		dftCounts = append(dftCounts, dft.String())
	}
	return fmt.Sprintf(`Statistics
	documents total: %d
	documents inbox: %d
	inbox tag: %s
	inbox tags: %s
	document file type count: %s
	character count: %d
	tag count %d
	correspondent count %d
	document type count: %d
	storage path count: %d
	current asn: %d
`,
		s.DocumentsTotal,
		s.DocumentsInbox,
		s.InboxTag,
		strings.Join(s.InboxTags, ", "),
		strings.Join(dftCounts, "\n"),
		s.CharacterCount,
		s.TagCount,
		s.CorrespondentCount,
		s.DocumentTypeCount,
		s.StoragePathCount,
		s.CurrentASN,
	)
}
