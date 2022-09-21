package agola

import (
	"strings"
	"time"

	"wecode.sorint.it/opensource/papagaio-api/types"
)

type AgolaCreateORGDto struct {
	ID         string               `json:"id"`
	Name       string               `json:"name"`
	Visibility types.VisibilityType `json:"visibility"`
}

type RemoteSourcesDto struct {
	Name string `json:"name"`
}

type CreateProjectRequestDto struct {
	Name             string               `json:"name"`
	ParentRef        string               `json:"parent_ref"`
	Visibility       types.VisibilityType `json:"visibility"`
	RemoteSourceName string               `json:"remote_source_name"`
	RepoPath         string               `json:"repo_path"`
}

type CreateProjectResponseDto struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Path             string               `json:"path"`
	ParentPath       string               `json:"parent_path"`
	Visibility       types.VisibilityType `json:"visibility"`
	GlobalVisibility string               `json:"global_visibility"`
}

type OrganizationMembersResponseDto struct {
	Members []MemberDto `json:"members"`
}

type MemberDto struct {
	User UserDto  `json:"user"`
	Role RoleType `json:"role"`
}

type RoleType string

const (
	Owner  RoleType = "owner"
	Member RoleType = "member"
)

type RunsDto struct {
	Number      uint64            `json:"number"`
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Phase       RunPhase          `json:"phase"`
	Result      RunResult         `json:"result"`

	TasksWaitingApproval []string `json:"tasks_waiting_approval"`

	EnqueueTime *time.Time `json:"enqueue_time"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
}

func (run *RunsDto) IsBranch() bool {
	return strings.Compare(run.Annotations["ref_type"], "branch") == 0
}

func (run *RunsDto) GetBranchName() string {
	return run.Annotations["branch"]
}

func (run *RunsDto) IsWebhookCreationTrigger() bool {
	return strings.Compare(run.Annotations["run_creation_trigger"], "webhook") == 0
}

type RunDto struct {
	Number      uint64            `json:"number"`
	Name        string            `json:"name"`
	Annotations map[string]string `json:"annotations"`
	Phase       RunPhase          `json:"phase"`
	Result      RunResult         `json:"result"`
	SetupErrors []string          `json:"setup_errors"`
	Stopping    bool              `json:"stopping"`

	Tasks                map[string]*TaskDto `json:"tasks"`
	TasksWaitingApproval []string            `json:"tasks_waiting_approval"`

	EnqueueTime *time.Time `json:"enqueue_time"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`

	CanRestartFromScratch     bool `json:"can_restart_from_scratch"`
	CanRestartFromFailedTasks bool `json:"can_restart_from_failed_tasks"`
}

type RunPhase string

const (
	RunPhaseSetupError RunPhase = "setuperror"
	RunPhaseQueued     RunPhase = "queued"
	RunPhaseCancelled  RunPhase = "cancelled"
	RunPhaseRunning    RunPhase = "running"
	RunPhaseFinished   RunPhase = "finished"
)

type RunResult string

const (
	RunResultUnknown RunResult = "unknown"
	RunResultStopped RunResult = "stopped"
	RunResultSuccess RunResult = "success"
	RunResultFailed  RunResult = "failed"
)

func (run *RunDto) IsWebhookCreationTrigger() bool {
	return strings.Compare(run.Annotations["run_creation_trigger"], "webhook") == 0
}

func (run *RunDto) GetBranchName() string {
	return run.Annotations["branch"]
}

func (run *RunDto) GetCommitSha() string {
	return run.Annotations["commit_sha"]
}

func (run *RunDto) IsBranch() bool {
	return strings.Compare(run.Annotations["ref_type"], "branch") == 0
}

type TaskDto struct {
	ID         string                     `json:"id"`
	Name       string                     `json:"name"`
	Status     RunTaskStatus              `json:"status"`
	Timedout   bool                       `json:"timedout"`
	Containers []RunTaskResponseContainer `json:"containers"`

	WaitingApproval     bool              `json:"waiting_approval"`
	Approved            bool              `json:"approved"`
	ApprovalAnnotations map[string]string `json:"approval_annotations"`

	SetupStep *RunTaskResponseSetupStep `json:"setup_step"`
	Steps     []*RunTaskResponseStep    `json:"steps"`

	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	TaskTimeoutInterval time.Duration `json:"task_timeout_interval"`
}

type RunTaskResponseContainer struct {
	Image string `json:"image"`
}

type RunTaskResponseSetupStep struct {
	Phase ExecutorTaskPhase `json:"phase"`
	Name  string            `json:"name"`

	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type RunTaskResponseStep struct {
	Phase   ExecutorTaskPhase `json:"phase"`
	Type    string            `json:"type"`
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Shell   string            `json:"shell"`

	ExitStatus *int `json:"exit_status"`

	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	LogArchived bool `json:"log_archived"`
}

type RunConfigTaskDepend struct {
	TaskID     string                         `json:"task_id,omitempty"`
	Conditions []RunConfigTaskDependCondition `json:"conditions,omitempty"`
}

type RunConfigTaskDependCondition string

const (
	RunConfigTaskDependConditionOnSuccess RunConfigTaskDependCondition = "on_success"
	RunConfigTaskDependConditionOnFailure RunConfigTaskDependCondition = "on_failure"
	RunConfigTaskDependConditionOnSkipped RunConfigTaskDependCondition = "on_skipped"
)

type RunTaskStatus string

const (
	RunTaskStatusNotStarted RunTaskStatus = "notstarted"
	RunTaskStatusSkipped    RunTaskStatus = "skipped"
	RunTaskStatusCancelled  RunTaskStatus = "cancelled"
	RunTaskStatusRunning    RunTaskStatus = "running"
	RunTaskStatusStopped    RunTaskStatus = "stopped"
	RunTaskStatusSuccess    RunTaskStatus = "success"
	RunTaskStatusFailed     RunTaskStatus = "failed"
)

type RunTaskStep struct {
	Phase      ExecutorTaskPhase `json:"phase"`
	Name       string            `json:"name"`
	LogPhase   RunTaskFetchPhase `json:"log_phase"`
	ExitStatus int               `json:"exit_status"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
}

type ExecutorTaskPhase string

const (
	ExecutorTaskPhaseNotStarted ExecutorTaskPhase = "notstarted"
	ExecutorTaskPhaseCancelled  ExecutorTaskPhase = "cancelled"
	ExecutorTaskPhaseRunning    ExecutorTaskPhase = "running"
	ExecutorTaskPhaseStopped    ExecutorTaskPhase = "stopped"
	ExecutorTaskPhaseSuccess    ExecutorTaskPhase = "success"
	ExecutorTaskPhaseFailed     ExecutorTaskPhase = "failed"
)

type RunTaskFetchPhase string

const (
	RunTaskFetchPhaseNotStarted RunTaskFetchPhase = "notstarted"
	RunTaskFetchPhaseFinished   RunTaskFetchPhase = "finished"
)

type RemoteSourceDto struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	AuthType            string `json:"auth_type"`
	RegistrationEnabled bool   `json:"registration_enabled"`
	LoginEnabled        bool   `json:"login_enabled"`
}

type UserDto struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type OrganizationDto struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type UserOrgDto struct {
	Role         RoleType        `json:"role"`
	Organization OrganizationDto `json:"organization"`
}

type ProjectDto struct {
	ID                 string `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	Path               string `json:"path,omitempty"`
	ParentPath         string `json:"parent_path,omitempty"`
	Visibility         string `json:"visibility,omitempty"`
	GlobalVisibility   string `json:"global_visibility,omitempty"`
	PassVarsToForkedPR bool   `json:"pass_vars_to_forked_pr,omitempty"`
	DefaultBranch      string `json:"default_branch,omitempty"`
}

type TokenRequestDto struct {
	TokenName string `json:"token_name"`
}

type TokenResponseDto struct {
	Token string `json:"token"`
}

type CreateRemoteSourceRequestDto struct {
	Name                string `json:"name"`
	APIURL              string `json:"apiurl"`
	Type                string `json:"type"`
	AuthType            string `json:"auth_type"`
	SkipVerify          bool   `json:"skip_verify"`
	Oauth2ClientID      string `json:"oauth_2_client_id"`
	Oauth2ClientSecret  string `json:"oauth_2_client_secret"`
	SSHHostKey          string `json:"ssh_host_key"`
	SkipSSHHostKeyCheck bool   `json:"skip_ssh_host_key_check"`
	RegistrationEnabled *bool  `json:"registration_enabled"`
	LoginEnabled        *bool  `json:"login_enabled"`
}

type ProjectGroupDto struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	ParentPath       string `json:"parent_path"`
	Visibility       string `json:"visibility"`
	GlobalVisibility string `json:"global_visibility"`
}
