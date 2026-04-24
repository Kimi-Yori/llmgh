package main

import "fmt"

func cmdStatus(args []string) error {
	client, err := NewClient()
	if err != nil {
		return err
	}

	owner, repo, _, repoErr := resolveRepo(args)

	if owner != "" && repo != "" {
		data, err := client.Get(fmt.Sprintf("/repos/%s/%s", owner, repo))
		if err != nil {
			return err
		}
		tsv("repo", owner+"/"+repo,
			"default_branch="+str(data["default_branch"]),
			"private="+str(data["private"]),
			"stars="+str(data["stargazers_count"]),
			"forks="+str(data["forks_count"]),
		)
	} else if repoErr != nil {
		tsv("repo", "none", repoErr.Error())
	}

	user := client.AuthUser()
	if user != "" {
		tsv("auth", "ok", "user="+user)
	} else if client.token == "" {
		tsv("auth", "none", "hint=set LLMGH_TOKEN or GH_TOKEN")
	} else {
		tsv("auth", "fail", "token present but invalid")
	}

	rateData, err := client.GetRateLimit()
	if err == nil {
		if resources, ok := rateData["resources"].(map[string]any); ok {
			if core, ok := resources["core"].(map[string]any); ok {
				tsv("rate", "core",
					"remaining="+str(core["remaining"]),
					"limit="+str(core["limit"]),
					"reset="+str(core["reset"]),
				)
			}
		}
	}

	return nil
}
