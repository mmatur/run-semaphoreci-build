package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/containous/flaeg"
	"github.com/ldez/go-semaphoreci/v1"
	"github.com/mmatur/run-semaphoreci-build/meta"
	"github.com/mmatur/run-semaphoreci-build/types"
	"github.com/ogier/pflag"
)

func main() {
	defaultCfg := &types.Config{}
	defaultPointerCfg := &types.Config{}

	rootCmd := &flaeg.Command{
		Name:                  "run-semaphoreci-build",
		Description:           "Run semaphoreci build",
		Config:                defaultCfg,
		DefaultPointersConfig: defaultPointerCfg,
		Run: func() error {
			return rootRun(defaultCfg)
		},
	}

	flag := flaeg.New(rootCmd, os.Args[1:])

	// version

	versionCmd := &flaeg.Command{
		Name:                  "version",
		Description:           "Display the version.",
		Config:                &types.NoOption{},
		DefaultPointersConfig: &types.NoOption{},
		Run: func() error {
			meta.DisplayVersion()
			return nil
		},
	}

	flag.AddCommand(versionCmd)

	// Run command

	if err := flag.Run(); err != nil && err != pflag.ErrHelp {
		log.Printf("Error: %v\n", err)
	}
}

func rootRun(config *types.Config) error {
	if config.SHA == "" {
		config.SHA = os.Getenv("GITHUB_SHA")
	}

	if err := validate(config); err != nil {
		return err
	}

	if config.TagEvent && !strings.Contains(os.Getenv("GITHUB_REF"), "refs/tags") {
		return nil
	}

	transport := v1.TokenTransport{
		Token: os.Getenv("SEMAPHORECI_TOKEN"),
	}

	client := v1.NewClient(transport.Client())

	project, err := findProject(client, config.Owner, config.Project)
	if err != nil {
		return err
	}

	branchID, err := findBranchID(client, project, config.Branch, config.SHA)
	if err != nil {
		return err
	}

	return launchBuild(client, project, branchID, config.SHA)
}

func findProject(client *v1.Client, ownerName string, projectName string) (v1.Project, error) {
	projects, _, err := client.Projects.Get()
	if err != nil {
		return v1.Project{}, err
	}

	for _, p := range projects {
		if p.Owner == ownerName && p.Name == projectName {
			return p, nil
		}
	}

	return v1.Project{}, fmt.Errorf("no project found for owner=%q and project=%q", ownerName, projectName)
}

func findBranchID(client *v1.Client, project v1.Project, branchName string, SHA string) (int, error) {
	branches, _, err := client.Branch.GetByProject(project.HashID)
	if err != nil {
		return 0, err
	}

	for _, b := range branches {
		if b.Name == branchName {
			return b.ID, nil
		}

		histories, _, err := client.Branch.GetHistory(project.HashID, b.ID, &v1.BranchHistoryOptions{})
		if err != nil || histories == nil {
			return 0, nil
		}

		for _, build := range histories.Builds {
			if build.Commit.ID == SHA {
				return b.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("no branch found for branch name %q or commit %q", branchName, SHA)
}

func launchBuild(client *v1.Client, project v1.Project, branchID int, sha string) error {
	information, err := client.Builds.Launch(project.HashID, branchID, sha)
	log.Println(information)
	return err
}

func validate(config *types.Config) error {
	if err := required(config.Owner, "owner"); err != nil {
		return err
	}

	if err := required(config.Project, "project"); err != nil {
		return err
	}

	return required(config.SHA, "sha")
}

func required(field string, fieldName string) error {
	if len(field) == 0 {
		return fmt.Errorf("option %s is mandatory", fieldName)
	}
	return nil
}
