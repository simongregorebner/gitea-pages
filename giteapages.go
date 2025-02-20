package giteapages

import (
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(GiteaPagesModule{})
	httpcaddyfile.RegisterHandlerDirective("gitea-pages", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var module GiteaPagesModule
	err := module.UnmarshalCaddyfile(h.Dispenser)
	return module, err
}

// GiteaagesModule implements gitea plugin.
type GiteaPagesModule struct {
	Logger                 *zap.Logger   `json:"-"`
	GiteaClient            *gitea.Client `json:"-"`
	Server                 string        `json:"server,omitempty"`
	Token                  string        `json:"token,omitempty"`
	PagesBranch            string        `json:"pages_branch,omitempty"`
	PostfixPagesRepository string        `json:"postfix_pages_repository,omitempty"`
	URLScheme              string        `json:"url_scheme,omitempty"`
}

// CaddyModule returns the Caddy module information.
func (GiteaPagesModule) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.gitea",
		New: func() caddy.Module { return new(GiteaPagesModule) },
	}
}

// Provision provisions gitea client.
func (module *GiteaPagesModule) Provision(ctx caddy.Context) error {
	var err error
	module.Logger = ctx.Logger()
	module.GiteaClient, err = gitea.NewClient(module.Server, gitea.SetToken(module.Token), gitea.SetGiteaVersion(""))
	return err
}

// UnmarshalCaddyfile unmarshals a Caddyfile.
func (module *GiteaPagesModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for n := d.Nesting(); d.NextBlock(n); {
			switch d.Val() {
			case "server":
				d.Args(&module.Server)
			case "token":
				d.Args(&module.Token)
			case "pages_branch":
				d.Args(&module.PagesBranch)
			case "postfix_pages_repository":
				d.Args(&module.PostfixPagesRepository)
			case "url_scheme":
				d.Args(&module.URLScheme)
			}
		}

		// Set defaults
		if module.PagesBranch == "" {
			module.PagesBranch = "gitea-pages"
		}
		if module.PostfixPagesRepository == "" {
			module.PostfixPagesRepository = "gitea.io"
		}

		if module.URLScheme == "" {
			module.URLScheme = "simple"
		}
		// Only accept simple and classic option
		switch module.URLScheme {
		case "simple", "classic":
		default:
			return errors.New("Invalid URL scheme: " + module.URLScheme)
		}

	}
	return nil
}

// ServeHTTP performs gitea content fetcher.
func (module GiteaPagesModule) ServeHTTP(writer http.ResponseWriter, request *http.Request, _ caddyhttp.Handler) error {

	var organization, repository, path string
	if module.URLScheme == "simple" {
		fmt.Println("SIMPLE")
		// "Simple" URL case - we expect the organization and repository in the URL
		// The URL/path looks like http(s)://<giteaserver>[:<port>]/<organization>/<repository>[/<filepath>]

		// Remove a potential "/" prefix and trailing "/" -  then split up the path
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/"), "/"), "/")

		length := len(parts)
		if length <= 1 {
			return caddyhttp.Error(http.StatusNotFound, fs.ErrNotExist)
		} else if length == 2 {
			organization = parts[0]
			repository = parts[1]
			path = "index.html" // there is no file/path specified
		} else {
			organization = parts[0]
			repository = parts[1]
			path = strings.Join(parts[2:], "/")
		}
	} else {
		// "Classic" URL/host scheme
		// The URL/path looks like http(s)://<organization>.<giteaserver>[:<port>]/<repository>/[/<filepath>]

		// extract the organization from the hostname
		organization = strings.TrimRight(request.Host, ".")

		// Remove a potential "/" prefix and trailing "/" -  then split up the path
		path = strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/"), "/")

		if path == "" {
			// Case http(s)://<organization>.<giteaserver>[:<port>]
			// (Try to) Use of the pages repository in the specified organization

			// We use github.com conventions: <organization>.github.io
			repository = organization + "." + module.PostfixPagesRepository
			path = "index.html"

		} else {
			parts := strings.Split(path, "/")
			// The part[0] can now be a repository name or already the filepath (organization wide pages repo)
			// Because we checked for "" before we know that we at least have one part

			// Check if parts[0] identifies a repo in the organization
			// TODO cache this result for some time part[0] == repository(true/false)
			if module.repoBranchExists(organization, parts[0], module.PagesBranch) {
				// Its a specifig repo inside the organization
				repository = parts[0]
				if len(parts) == 1 {
					path = "index.html"
				} else {
					path = strings.Join(parts[1:], "/")
				}
			} else {
				// (Try to) Use of the pages repository in the specified organization
				// We use github.com conventions: <organization>.github.io
				repository = organization + "." + module.PostfixPagesRepository
				path = strings.Join(parts[0:], "/")
			}
		}
	}

	// Handle request
	content, err := module.getFile(organization, repository, module.PagesBranch, path)
	if err != nil {
		return caddyhttp.Error(http.StatusNotFound, err)
	}

	// Try to determine mime type based on extenstion of file
	parts := strings.Split(request.URL.Path, ".")
	if len(parts) > 1 {
		extension := parts[len(parts)-1] // get file extension
		writer.Header().Add("Content-Type", mime.TypeByExtension("."+extension))
	}
	_, err = writer.Write(content)
	// _, err = io.Copy(writer, content)
	return err
}

// Retrieve specific file from gitea server
func (module GiteaPagesModule) getFile(organization, repository, branch, path string) ([]byte, error) {

	module.Logger.Info(fmt.Sprintf("Retrieve file - owner: %s repo: %s filepath: %s branch: %s", organization, repository, path, branch))

	content, _, err := module.GiteaClient.GetFile(organization, repository, branch, path)
	return content, err
}

// Check if the repo has a specific branch
func (module GiteaPagesModule) repoBranchExists(organization, repository, branch string) bool {
	branchInfo, _, err := module.GiteaClient.GetRepoBranch(organization, repository, branch)
	if err != nil {
		return false
	}
	return branchInfo.Name == branch
}

// Check if a repo has a specific topic assigned
func (module GiteaPagesModule) topicExists(organization, repository, topic string) bool {
	topics, _, err := module.GiteaClient.ListRepoTopics(organization, repository, gitea.ListRepoTopicsOptions{})
	if err != nil {
		return false
	}

	for _, topicName := range topics {
		if topicName == topic {
			return true
		}
	}
	return false
}

// Interface guards
var (
	_ caddy.Provisioner           = (*GiteaPagesModule)(nil)
	_ caddyhttp.MiddlewareHandler = (*GiteaPagesModule)(nil)
	_ caddyfile.Unmarshaler       = (*GiteaPagesModule)(nil)
)
