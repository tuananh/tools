package graph

import (
	"context"
	"fmt"
	"strings"
	"path/filepath"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/gptscript-ai/go-gptscript"
)

func ListThreadMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string) ([]models.Postable, error) {
	// Fetch messages inside a thread
	result, err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Posts().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list group mailbox messages: %w", err)
	}

	return result.GetValue(), nil
}

func ListGroupThreads(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, start, end string, limit int) ([]models.ConversationThreadable, error) {
	queryParams := &groups.ItemThreadsRequestBuilderGetQueryParameters{
		Orderby: []string{"lastDeliveredDateTime DESC"},
	}

	if limit > 0 {
		queryParams.Top = util.Ptr(int32(limit))
	}

	var filters []string
	if start != "" {
		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime ge %s", start))
	}
	if end != "" {
		filters = append(filters, fmt.Sprintf("lastDeliveredDateTime le %s", end))
	}

	if len(filters) > 0 {
		queryParams.Filter = util.Ptr(strings.Join(filters, " and "))
	}

	// Fetch messages from the group mailbox
	result, err := client.Groups().ByGroupId(groupID).Threads().Get(ctx, &groups.ItemThreadsRequestBuilderGetRequestConfiguration{
		QueryParameters: queryParams,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list group mailbox messages: %w", err)
	}

	return result.GetValue(), nil
}

// ListGroups retrieves all Microsoft 365 groups the authenticated user has access to
func ListGroups(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.Groupable, error) {

	// Fetch groups where the user is a member
	result, err := client.Me().MemberOf().Get(ctx, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}

	// Filter for groups that have a mailbox (mailEnabled == true)
	var accessibleGroups []models.Groupable
	for _, group := range result.GetValue() {
		if g, ok := group.(models.Groupable); ok {
			if g.GetMailEnabled() != nil && *g.GetMailEnabled() {
				accessibleGroups = append(accessibleGroups, g)
			}
		}
	}

	return accessibleGroups, nil
}

func getGroup(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID string) (models.Groupable, error) {
	groups, err := client.Groups().ByGroupId(groupID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return groups, nil
}


func CreateGroupThreadMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID string, info DraftInfo) (models.ConversationThreadable, error) {

	for _, file := range info.Attachments {
		if file == "" {
			return nil, fmt.Errorf("attachment file path cannot be empty")
		}
	}

	requestBody := models.NewConversationThread()
	requestBody.SetTopic(util.Ptr(info.Subject)) 

	post := models.NewPost()
	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE)) 
	body.SetContent(util.Ptr(info.Body)) 
	post.SetBody(body)

	if len(info.Recipients) > 0 {
		post.SetNewParticipants(emailAddressesToRecipientable(info.Recipients))
	}

	// models.Post() doesn't support cc and bcc
	
	if len(info.Attachments) > 0 {
		attachments, err := setAttachments(ctx, info.Attachments)
		if err != nil {
			return nil, fmt.Errorf("failed to attach files to group thread message post: %w", err)
		}
		post.SetAttachments(attachments)
	}
	posts := []models.Postable {
		post,
	}

	requestBody.SetPosts(posts)

	threads, err := client.Groups().ByGroupId(groupID).Threads().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create group thread message: %w", err)
	}

	return threads, nil
}



func ReplyToGroupThreadMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string, info DraftInfo) ( error) {
	for _, file := range info.Attachments {
		if file == "" {
			return fmt.Errorf("attachment file path cannot be empty")
		}
	}
	
	requestBody := groups.NewItemConversationsItemThreadsItemReplyPostRequestBody()
	post := models.NewPost()
	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE)) 
	body.SetContent(util.Ptr(info.Body)) 
	post.SetBody(body)

	if len(info.Recipients) > 0 {
		post.SetNewParticipants(emailAddressesToRecipientable(info.Recipients))
	}

	// models.Post() doesn't support cc and bcc
	
	if len(info.Attachments) > 0 {
		attachments, err := setAttachments(ctx, info.Attachments)
		if err != nil {
			return fmt.Errorf("failed to attach files to group thread message post: %w", err)
		}
		post.SetAttachments(attachments)
	}
	requestBody.SetPost(post)

	err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Reply().Post(ctx, requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to reply to group thread message %s: %w", threadID, err)
	}

	return nil
}

func setAttachments(ctx context.Context, attachment_filenames []string) ([]models.Attachmentable, error) {
	attachments := []models.Attachmentable{}
	gsClient, err := gptscript.NewGPTScript()
	if err != nil {
		return nil, fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	for _, filename := range attachment_filenames {
		attachment := models.NewFileAttachment()
		attachment.SetName(util.Ptr(filename)) 

		data, err := gsClient.ReadFileInWorkspace(ctx, filepath.Join("files", filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment file %s from workspace: %v", filename, err)
		}

		attachment.SetContentBytes(data) 
		attachments = append(attachments, attachment)
	}

	return attachments, nil

}

func DeleteGroupThread(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, groupID, threadID string) error {
	err := client.Groups().ByGroupId(groupID).Threads().ByConversationThreadId(threadID).Delete(ctx, nil)
	if err != nil {
		fmt.Errorf("failed to delete group thread: %w", err)
		return err
	}
	return nil
}
