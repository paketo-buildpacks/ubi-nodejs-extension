package structs

type DuringBuildPermissions struct {
	CNB_USER_ID, CNB_GROUP_ID int
}

type BuildDockerfileProps struct {
	NODEJS_VERSION            uint64
	CNB_USER_ID, CNB_GROUP_ID int
	CNB_STACK_ID, PACKAGES    string
	SET_SYMLINKS              string
	ENABLE_NODEJS_MODULE      bool
}

type RunDockerfileProps struct {
	Source string
}
