package dao

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type githubApi struct {
}

func NewGithubApi() *githubApi {
	return &githubApi{}
}

var githubInfoApi = "https://api.github.com/search/repositories?q=repo:"

type GithubInfoResp struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		Owner    struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"owner"`
		HTMLURL          string      `json:"html_url"`
		Description      string      `json:"description"`
		Fork             bool        `json:"fork"`
		URL              string      `json:"url"`
		ForksURL         string      `json:"forks_url"`
		KeysURL          string      `json:"keys_url"`
		CollaboratorsURL string      `json:"collaborators_url"`
		TeamsURL         string      `json:"teams_url"`
		HooksURL         string      `json:"hooks_url"`
		IssueEventsURL   string      `json:"issue_events_url"`
		EventsURL        string      `json:"events_url"`
		AssigneesURL     string      `json:"assignees_url"`
		BranchesURL      string      `json:"branches_url"`
		TagsURL          string      `json:"tags_url"`
		BlobsURL         string      `json:"blobs_url"`
		GitTagsURL       string      `json:"git_tags_url"`
		GitRefsURL       string      `json:"git_refs_url"`
		TreesURL         string      `json:"trees_url"`
		StatusesURL      string      `json:"statuses_url"`
		LanguagesURL     string      `json:"languages_url"`
		StargazersURL    string      `json:"stargazers_url"`
		ContributorsURL  string      `json:"contributors_url"`
		SubscribersURL   string      `json:"subscribers_url"`
		SubscriptionURL  string      `json:"subscription_url"`
		CommitsURL       string      `json:"commits_url"`
		GitCommitsURL    string      `json:"git_commits_url"`
		CommentsURL      string      `json:"comments_url"`
		IssueCommentURL  string      `json:"issue_comment_url"`
		ContentsURL      string      `json:"contents_url"`
		CompareURL       string      `json:"compare_url"`
		MergesURL        string      `json:"merges_url"`
		ArchiveURL       string      `json:"archive_url"`
		DownloadsURL     string      `json:"downloads_url"`
		IssuesURL        string      `json:"issues_url"`
		PullsURL         string      `json:"pulls_url"`
		MilestonesURL    string      `json:"milestones_url"`
		NotificationsURL string      `json:"notifications_url"`
		LabelsURL        string      `json:"labels_url"`
		ReleasesURL      string      `json:"releases_url"`
		DeploymentsURL   string      `json:"deployments_url"`
		CreatedAt        time.Time   `json:"created_at"`
		UpdatedAt        time.Time   `json:"updated_at"`
		PushedAt         time.Time   `json:"pushed_at"`
		GitURL           string      `json:"git_url"`
		SSHURL           string      `json:"ssh_url"`
		CloneURL         string      `json:"clone_url"`
		SvnURL           string      `json:"svn_url"`
		Homepage         string      `json:"homepage"`
		Size             int         `json:"size"`
		StargazersCount  int         `json:"stargazers_count"`
		WatchersCount    int         `json:"watchers_count"`
		Language         string      `json:"language"`
		HasIssues        bool        `json:"has_issues"`
		HasProjects      bool        `json:"has_projects"`
		HasDownloads     bool        `json:"has_downloads"`
		HasWiki          bool        `json:"has_wiki"`
		HasPages         bool        `json:"has_pages"`
		ForksCount       int         `json:"forks_count"`
		MirrorURL        interface{} `json:"mirror_url"`
		Archived         bool        `json:"archived"`
		Disabled         bool        `json:"disabled"`
		OpenIssuesCount  int         `json:"open_issues_count"`
		License          struct {
			Key    string `json:"key"`
			Name   string `json:"name"`
			SpdxID string `json:"spdx_id"`
			URL    string `json:"url"`
			NodeID string `json:"node_id"`
		} `json:"license"`
		Forks         int     `json:"forks"`
		OpenIssues    int     `json:"open_issues"`
		Watchers      int     `json:"watchers"`
		DefaultBranch string  `json:"default_branch"`
		Score         float64 `json:"score"`
	} `json:"items"`
}

func (*githubApi) Info(name string) (output mysql.Awesome, err error) {
	var info *resty.Response
	info, err = mus.JsonRestyClient.R().Get(githubInfoApi + name)
	if err != nil {
		return
	}
	var resp GithubInfoResp
	err = json.Unmarshal(info.Body(), &resp)
	if err != nil {
		return
	}

	if len(resp.Items) == 0 {
		err = errors.New("length is error")
		return
	}

	if resp.Items[0].FullName != name {
		err = errors.New("project is error")
		return
	}

	output = mysql.Awesome{
		Name:           name,
		GitName:        resp.Items[0].Name,
		OwnerAvatarUrl: resp.Items[0].Owner.AvatarURL,
		HtmlUrl:        resp.Items[0].HTMLURL,
		GitDescription: resp.Items[0].Description,
		GitCreatedAt:   resp.Items[0].CreatedAt,
		GitUpdatedAt:   resp.Items[0].UpdatedAt,
		GitUrl:         resp.Items[0].GitURL,
		SshUrl:         resp.Items[0].SSHURL,
		CloneUrl:       resp.Items[0].CloneURL,
		HomePage:       resp.Items[0].Homepage,
		StarCount:      resp.Items[0].StargazersCount,
		WatcherCount:   resp.Items[0].WatchersCount,
		Language:       resp.Items[0].Language,
		ForkCount:      resp.Items[0].ForksCount,
		LicenseKey:     resp.Items[0].License.Key,
		LicenseName:    resp.Items[0].License.Name,
		LicenseUrl:     resp.Items[0].License.URL,
	}
	return
}

func (*githubApi) Update(param mysql.Awesome, version int) (err error) {
	err = mus.Db.Model(mysql.Awesome{}).Where("name = ?", param.Name).UpdateColumns(param).Error
	if err != nil {
		return
	}

	if version == 0 {
		mus.Db.Model(mysql.Awesome{}).Where("name = ?", param.Name).Updates(map[string]interface{}{
			"version": gorm.Expr("version+?", 1),
		})
	} else {
		mus.Db.Model(mysql.Awesome{}).Where("name = ?", param.Name).Updates(map[string]interface{}{
			"version": version,
		})
	}

	return

}

func (g *githubApi) All() (err error) {
	// 取出最大的
	var maxVersion mysql.Awesome
	mus.Db.Order("version desc").Limit(1).Find(&maxVersion)
	version := maxVersion.Version + 1

	for {
		var arr []mysql.Awesome
		err = mus.Db.Model(mysql.Awesome{}).Where("version < ?", version).Limit(20).Find(&arr).Error
		if err != nil || len(arr) == 0 {
			return nil
		}
		for _, value := range arr {
			var mysqlInfo mysql.Awesome
			// 避免出问题，下次继续循环
			mus.Db.Model(mysql.Awesome{}).Where("name = ?", value.Name).Updates(map[string]interface{}{
				"version": version,
			})
			mysqlInfo, err = g.Info(value.Name)
			if err != nil {
				mus.Logger.Error("api info err", zap.String("name", value.Name), zap.Error(err))
				continue
			}
			err = g.Update(mysqlInfo, version)
			if err != nil {
				mus.Logger.Error("api info err", zap.String("name", value.Name), zap.Error(err))
			}
		}
	}

}
