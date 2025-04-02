package commands

import (
	"context"
	"fmt"

	graph "github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
)

func DeleteGroupThread(ctx context.Context, groupID, threadID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteGroupThread(ctx, c, groupID, threadID); err != nil {
		return err
	}

	fmt.Printf("Group thread %s deleted successfully\n", threadID)
	return nil
}
