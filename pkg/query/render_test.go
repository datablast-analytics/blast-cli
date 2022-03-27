package query

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderer_RenderQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		args    map[string]string
		want    string
		wantErr bool
	}{
		{
			name:  "simple render for ds",
			query: "set analysis_end_date = '{{ ds }}'::date;",
			args: map[string]string{
				"ds": "2022-02-03",
			},
			want: "set analysis_end_date = '2022-02-03'::date;",
		},
		{
			name:  "multiple variables",
			query: "set analysis_end_date = '{{ ds }}'::date and '{{testVar}}' == 'testvar' and another date {{    ds }} - {{ someMissingVariable }};",
			args: map[string]string{
				"ds":      "2022-02-03",
				"testVar": "testvar",
			},
			want: "set analysis_end_date = '2022-02-03'::date and 'testvar' == 'testvar' and another date 2022-02-03 - {{ someMissingVariable }};",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			receiver := Renderer{}
			got, err := receiver.RenderQuery(tt.query, tt.args)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
