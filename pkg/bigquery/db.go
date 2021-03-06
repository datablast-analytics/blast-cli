package bigquery

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/datablast-analytics/blast-cli/pkg/query"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var scopes = []string{
	bigquery.Scope,
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/drive",
}

type DB struct {
	client *bigquery.Client
}

func NewDB(c *Config) (*DB, error) {
	client, err := bigquery.NewClient(
		context.Background(),
		c.ProjectID,
		option.WithCredentialsFile(c.CredentialsFilePath),
		option.WithScopes(scopes...),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create bigquery client")
	}

	if c.Location != "" {
		client.Location = c.Location
	}

	return &DB{
		client: client,
	}, nil
}

func (d DB) IsValid(ctx context.Context, query *query.Query) (bool, error) {
	q := d.client.Query(query.ToDryRunQuery())
	q.DryRun = true

	job, err := q.Run(ctx)
	if err != nil {
		var googleError *googleapi.Error
		if !errors.As(err, &googleError) {
			return false, err
		}

		if googleError.Code == 404 {
			return false, fmt.Errorf("%s", googleError.Message)
		}

		return false, googleError
	}

	status := job.LastStatus()
	if err := status.Err(); err != nil {
		return false, err
	}

	return true, nil
}
