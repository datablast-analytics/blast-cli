package lint

import (
	"os"
	"testing"

	"github.com/datablast-analytics/blast-cli/pkg/pipeline"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var noIssues = make([]*Issue, 0)

func TestEnsureTaskNameIsNotEmpty(t *testing.T) {
	t.Parallel()

	taskWithEmptyName := pipeline.Task{
		Name: "",
	}

	type args struct {
		pipeline *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "all tasks have names, no error",
			args: args{
				pipeline: &pipeline.Pipeline{
					Name: "test",
					Tasks: []*pipeline.Task{
						{
							Name: "task1",
						},
						{
							Name: "task2",
						},
					},
				},
			},
			want:    make([]*Issue, 0),
			wantErr: false,
		},
		{
			name: "tasks with missing name are reported",
			args: args{
				pipeline: &pipeline.Pipeline{
					Name: "test",
					Tasks: []*pipeline.Task{
						{
							Name: "task1",
						},
						&taskWithEmptyName,
						{
							Name: "some-other-task",
						},
						{
							Name: "task name with spaces",
						},
					},
				},
			},
			want: []*Issue{
				{
					Task:        &taskWithEmptyName,
					Description: taskNameMustExist,
				},
				{
					Task: &pipeline.Task{
						Name: "task name with spaces",
					},
					Description: taskNameMustBeAlphanumeric,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsureTaskNameIsValid(tt.args.pipeline)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureExecutableFileIsValid(t *testing.T) {
	t.Parallel()

	type args struct {
		setupFilesystem func(t *testing.T, fs afero.Fs)
		pipeline        pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "comment task is skipped",
			args: args{
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.CommentTask,
							},
						},
					},
				},
			},
			want: noIssues,
		},
		{
			name: "task with no executable is skipped",
			args: args{
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
						},
					},
				},
			},
			want: noIssues,
		},
		{
			name: "task with no executable is reported for python files",
			args: args{
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Type: taskTypePython,
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						Type: taskTypePython,
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
					},
					Description: executableFileCannotBeEmpty,
				},
			},
		},
		{
			name: "task with no executable is skipped",
			args: args{
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-file.sh",
								Path: "some-path.sh",
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
						ExecutableFile: pipeline.ExecutableFile{
							Name: "some-file.sh",
							Path: "some-path.sh",
						},
					},
					Description: executableFileDoesNotExist,
				},
			},
		},
		{
			name: "executable is a directory",
			args: args{
				setupFilesystem: func(t *testing.T, fs afero.Fs) {
					err := fs.MkdirAll("some-path/some-file", 0o644)
					require.NoError(t, err, "failed to create the in-memory directory")
				},
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-file",
								Path: "some-path/some-file",
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
						ExecutableFile: pipeline.ExecutableFile{
							Name: "some-file",
							Path: "some-path/some-file",
						},
					},
					Description: executableFileIsADirectory,
				},
			},
		},
		{
			name: "executable is an empty file",
			args: args{
				setupFilesystem: func(t *testing.T, fs afero.Fs) {
					fileName := "some-path/some-file.sh"
					file, err := fs.Create(fileName)
					require.NoError(t, err)
					err = fs.Chmod(fileName, 0o755)
					require.NoError(t, err)
					require.NoError(t, file.Close())
				},
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-file.sh",
								Path: "some-path/some-file.sh",
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
						ExecutableFile: pipeline.ExecutableFile{
							Name: "some-file.sh",
							Path: "some-path/some-file.sh",
						},
					},
					Description: executableFileIsEmpty,
				},
				{
					Task: &pipeline.Task{
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
						ExecutableFile: pipeline.ExecutableFile{
							Name: "some-file.sh",
							Path: "some-path/some-file.sh",
						},
					},
					Description: executableFileIsNotExecutable,
				},
			},
		},
		{
			name: "executable file has the wrong permissions",
			args: args{
				setupFilesystem: func(t *testing.T, fs afero.Fs) {
					fileName := "some-path/some-file.sh"
					file, err := fs.Create(fileName)
					require.NoError(t, err)
					err = fs.Chmod(fileName, os.FileMode(0o100))
					require.NoError(t, err)
					defer func() { require.NoError(t, file.Close()) }()

					_, err = file.Write([]byte("some content"))
					require.NoError(t, err)
				},
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-file.sh",
								Path: "some-path/some-file.sh",
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						DefinitionFile: pipeline.DefinitionFile{
							Type: pipeline.YamlTask,
						},
						ExecutableFile: pipeline.ExecutableFile{
							Name: "some-file.sh",
							Path: "some-path/some-file.sh",
						},
					},
					Description: executableFileIsNotExecutable,
				},
			},
		},
		{
			name: "all good for the executable, no issues found",
			args: args{
				setupFilesystem: func(t *testing.T, fs afero.Fs) {
					file, err := fs.Create("some-path/some-file.sh")
					require.NoError(t, err)
					defer func() { require.NoError(t, file.Close()) }()

					err = fs.Chmod("some-path/some-file.sh", 0o644)
					require.NoError(t, err)

					_, err = file.Write([]byte("some content"))
					require.NoError(t, err)

					file, err = fs.Create("some-path/some-other-file.sh")
					require.NoError(t, err)
					defer func() { require.NoError(t, file.Close()) }()

					err = fs.Chmod("some-path/some-other-file.sh", 0o644)
					require.NoError(t, err)

					_, err = file.Write([]byte("some other content"))
					require.NoError(t, err)
				},
				pipeline: pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-file.sh",
								Path: "some-path/some-file.sh",
							},
						},
						{
							DefinitionFile: pipeline.DefinitionFile{
								Type: pipeline.YamlTask,
							},
							ExecutableFile: pipeline.ExecutableFile{
								Name: "some-other-file.sh",
								Path: "some-path/some-other-file.sh",
							},
						},
					},
				},
			},
			want: noIssues,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			if tt.args.setupFilesystem != nil {
				tt.args.setupFilesystem(t, fs)
			}

			checker := EnsureExecutableFileIsValid(fs)

			got, err := checker(&tt.args.pipeline)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureDependencyExists(t *testing.T) {
	t.Parallel()

	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "empty pipeline works fine",
			args: args{
				p: &pipeline.Pipeline{},
			},
			want: noIssues,
		},
		{
			name: "pipeline with no dependency has no issues",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name: "task1",
						},
						{
							Name: "task2",
						},
						{
							Name: "task3",
						},
					},
				},
			},
			want: noIssues,
		},
		{
			name: "dependency on a non-existing task is caught",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name:      "task1",
							DependsOn: []string{},
						},
						{
							Name:      "task2",
							DependsOn: []string{"task1", "task3", "task5"},
						},
						{
							Name:      "task3",
							DependsOn: []string{"task1", "task4"},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						Name:      "task2",
						DependsOn: []string{"task1", "task3", "task5"},
					},
					Description: "Dependency 'task5' does not exist",
				},
				{
					Task: &pipeline.Task{
						Name:      "task3",
						DependsOn: []string{"task1", "task4"},
					},
					Description: "Dependency 'task4' does not exist",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsureDependencyExists(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsurePipelineScheduleIsValidCron(t *testing.T) {
	t.Parallel()

	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "empty schedule is skipped",
			args: args{
				p: &pipeline.Pipeline{
					Schedule: "",
				},
			},
			want: noIssues,
		},
		{
			name: "invalid schedule is reported",
			args: args{
				p: &pipeline.Pipeline{
					Schedule: "some random schedule",
				},
			},
			want: []*Issue{
				{
					Description: "Invalid cron schedule 'some random schedule'",
				},
			},
		},
		{
			name: "valid schedule passes the check",
			args: args{
				p: &pipeline.Pipeline{
					Schedule: "* * * 1 *",
				},
			},
			want: noIssues,
		},
		{
			name: "valid descriptor passes the check",
			args: args{
				p: &pipeline.Pipeline{
					Schedule: "@daily",
				},
			},
			want: noIssues,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsurePipelineScheduleIsValidCron(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureOnlyAcceptedTaskTypesAreThere(t *testing.T) {
	t.Parallel()

	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "task with empty type is skipped",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Type: "",
						},
					},
				},
			},
			want: noIssues,
		},
		{
			name: "task invalid type is flagged",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Type: "some.random.type",
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						Type: "some.random.type",
					},
					Description: "Invalid task type 'some.random.type'",
				},
			},
		},
		{
			name: "task with valid type is not flagged",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Type: "bq.sql",
						},
					},
				},
			},
			want: noIssues,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsureOnlyAcceptedTaskTypesAreThere(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsureTaskNameIsUnique(t *testing.T) {
	t.Parallel()
	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "empty name is skipped",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name: "",
						},
					},
				},
			},
			want:    noIssues,
			wantErr: false,
		},
		{
			name: "duplicates are reported",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name: "name1",
							DefinitionFile: pipeline.DefinitionFile{
								Path: "path1",
							},
						},
						{
							Name: "name2",
							DefinitionFile: pipeline.DefinitionFile{
								Path: "path2",
							},
						},
						{
							Name: "name1",
							DefinitionFile: pipeline.DefinitionFile{
								Path: "path3",
							},
						},
					},
				},
			},
			want: []*Issue{
				{
					Task: &pipeline.Task{
						Name: "name1",
						DefinitionFile: pipeline.DefinitionFile{
							Path: "path1",
						},
					},
					Description: "Task name 'name1' is not unique, please make sure all the task names are unique",
					Context:     []string{"path1", "path3"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsureTaskNameIsUnique(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsurePipelineNameIsValid(t *testing.T) {
	t.Parallel()
	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "empty pipeline name is reported",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name: "",
						},
					},
				},
			},
			want: []*Issue{
				{
					Description: pipelineNameCannotBeEmpty,
					Context:     nil,
				},
			},
			wantErr: false,
		},
		{
			name: "spaces are not accepted",
			args: args{
				p: &pipeline.Pipeline{
					Name: "some test pipeline",
				},
			},
			want: []*Issue{
				{
					Description: pipelineNameMustBeAlphanumeric,
					Context:     nil,
				},
			},
			wantErr: false,
		},
		{
			name: "valid pipeline name passes",
			args: args{
				p: &pipeline.Pipeline{
					Name: "test",
				},
			},
			want:    []*Issue{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsurePipelineNameIsValid(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestEnsurePipelineHasNoCycles(t *testing.T) {
	t.Parallel()
	type args struct {
		p *pipeline.Pipeline
	}
	tests := []struct {
		name    string
		args    args
		want    []*Issue
		wantErr bool
	}{
		{
			name: "cycles are detected",
			args: args{
				p: &pipeline.Pipeline{
					Tasks: []*pipeline.Task{
						{
							Name: "task0",
						},
						{
							Name:      "task1",
							DependsOn: []string{"task2", "task0"},
						},
						{
							Name:      "task2",
							DependsOn: []string{"task3"},
						},
						{
							Name:      "task3",
							DependsOn: []string{"task1"},
						},
						{
							Name: "task4",
						},
						{
							Name:      "task5",
							DependsOn: []string{"task4", "task1"},
						},
						{
							Name:      "task6",
							DependsOn: []string{"task4", "task6"},
						},
					},
				},
			},
			want: []*Issue{
				{
					Description: pipelineContainsCycle,
					Context: []string{
						"Task `task6` depends on itself",
					},
				},
				{
					Description: pipelineContainsCycle,
					Context: []string{
						"task3 ??? task1",
						"task2 ??? task3",
						"task1 ??? task2",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := EnsurePipelineHasNoCycles(tt.args.p)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, got)
		})
	}
}
