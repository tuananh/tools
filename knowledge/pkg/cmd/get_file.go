package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/obot-platform/tools/knowledge/pkg/index/types"
	"github.com/spf13/cobra"
)

type ClientGetFile struct {
	Client
	Dataset string `usage:"Target Dataset ID" short:"d"`
}

func (s *ClientGetFile) Customize(cmd *cobra.Command) {
	cmd.Use = "get-file <file-id|file-abs-path>"
	cmd.Short = "Get a file from a dataset"
	cmd.Args = cobra.ExactArgs(1)
}

func (s *ClientGetFile) Run(cmd *cobra.Command, args []string) error {
	if s.Dataset == "" {
		exitErr0(fmt.Errorf("no dataset specified"))
	}

	c, err := s.getClient(cmd.Context())
	if err != nil {
		return err
	}
	defer c.Close()

	fileRef := args[0]

	searchFile := types.File{
		Dataset: s.Dataset,
	}

	if strings.HasPrefix(fileRef, "/") {
		searchFile.AbsolutePath = fileRef
	} else if _, err := uuid.Parse(fileRef); err == nil {
		searchFile.ID = fileRef
	} else {
		finfo, err := os.Stat(fileRef)
		if err != nil {
			return fmt.Errorf("fileref is not a valid filepath or UUID - failed to stat relative path: %w", err)
		}
		if finfo.IsDir() {
			return fmt.Errorf("fileref is a directory, not a file")
		}
		searchFile.AbsolutePath, err = filepath.Abs(fileRef)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	file, err := c.FindFile(cmd.Context(), searchFile)
	if err != nil {
		if errors.Is(err, types.ErrDBFileNotFound) {
			fmt.Printf("File not found: %s\n", fileRef)
			return nil
		}
		return err
	}

	jsonOutput, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("failed to marshal file: %w", err)
	}

	fmt.Println(string(jsonOutput))

	return nil
}
