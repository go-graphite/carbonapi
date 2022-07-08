package github

import (
	"context"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type API struct {
	ctx          *context.Context
	githubClient *githubv4.Client
}

func NewAPI(ctx context.Context, pat string) *API {
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: pat})
	httpClient := oauth2.NewClient(ctx, src)
	githubClient := githubv4.NewClient(httpClient)

	return &API{
		ctx:          &ctx,
		githubClient: githubClient,
	}
}

type PullRequestComment struct {
	ID          githubv4.ID
	Body        string
	IsMinimized bool
	Author      struct {
		Login string
	}
}

func (a *API) ListPullRequestComments(repoOwner string, repoName string, pullRequestNo int) ([]PullRequestComment, error) {
	var q struct {
		Repository struct {
			PullRequest struct {
				Comments struct {
					Nodes    []PullRequestComment
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"comments(first: 100, after: $commentsCursor)"`
			} `graphql:"pullRequest(number: $pr)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":          githubv4.String(repoOwner),
		"name":           githubv4.String(repoName),
		"pr":             githubv4.Int(pullRequestNo),
		"commentsCursor": (*githubv4.String)(nil),
	}

	var allComments []PullRequestComment
	for {
		err := a.githubClient.Query(*a.ctx, &q, variables)
		if err != nil {
			return nil, err
		}

		allComments = append(allComments, q.Repository.PullRequest.Comments.Nodes...)

		if !q.Repository.PullRequest.Comments.PageInfo.HasNextPage {
			break
		}
		variables["commentsCursor"] = githubv4.NewString(q.Repository.PullRequest.Comments.PageInfo.EndCursor)
	}

	return allComments, nil
}

func (a *API) MinimizeComment(commentNodeID githubv4.ID, classifier githubv4.ReportedContentClassifiers) error {
	var m struct {
		MinimizeComment struct {
			MinimizedComment struct {
				IsMinimized bool
			}
		} `graphql:"minimizeComment(input: $input)"`
	}
	input := githubv4.MinimizeCommentInput{
		SubjectID:  commentNodeID,
		Classifier: classifier,
	}

	err := a.githubClient.Mutate(*a.ctx, &m, input, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *API) GetPullRequestNodeID(repoOwner string, repoName string, pullRequestNo int) (githubv4.ID, error) {
	var q struct {
		Repository struct {
			PullRequest struct {
				ID githubv4.ID
			} `graphql:"pullRequest(number: $pr)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(repoOwner),
		"name":  githubv4.String(repoName),
		"pr":    githubv4.Int(pullRequestNo),
	}

	err := a.githubClient.Query(*a.ctx, &q, variables)
	if err != nil {
		return "", err
	}

	return q.Repository.PullRequest.ID, nil
}

func (a *API) AddComment(pullRequestNodeID githubv4.ID, commentBody string) (githubv4.ID, error) {
	var m struct {
		AddComment struct {
			CommentEdge struct {
				Node struct {
					ID githubv4.ID
				}
			}
		} `graphql:"addComment(input: $input)"`
	}
	input := githubv4.AddCommentInput{
		SubjectID: pullRequestNodeID,
		Body:      githubv4.String(commentBody),
	}

	err := a.githubClient.Mutate(*a.ctx, &m, input, nil)
	if err != nil {
		return "", err
	}

	return m.AddComment.CommentEdge.Node.ID, nil
}
