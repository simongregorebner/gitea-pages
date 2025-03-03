# Gitea-Pages

<img src="doc/gitea-pages.svg" alt="gitea-pages-logo" width="15%" height="auto">



# Overview

This is an all-in-one pages server for [Gitea](https://gitea.com). It works and uses similar/same conventions as [Github Pages](https://pages.github.com).


You can operate it in 2 modes, either __simple__ (default) or __classic__ (similar to how Github Pages operates).

## Simple Mode

In __simple__ mode no special DNS setup is required and the access to the hosted sides are always according to the pattern 

__http(s)://&lt;your-server-hostname&gt;/&lt;organization&gt;/&lt;repository&gt;__ 

To create a pages site, just 
1. Create a `gitea-pages` branch in any repository of an organization
2. Put your content in this branch. (eg index.html) and push the branch

After the push the content will be available at: http(s)://&lt;your-server-hostname&gt;/&lt;organization&gt;/&lt;repository&gt;

## Classic Mode

In __classic__ mode the access to the pages goes according to these two patterns: 

__http(s)://&lt;organization&gt;.&lt;your-server-hostname&gt;/&lt;repository&gt;__

or

__http(s)://&lt;organization&gt;.&lt;your-server-hostname&gt;__ 

The latter url scheme with serves the content of the repo named __gitea-pages__ of the organization (with default settings, however you can also set it to github like convention or  __&lt;organization&gt;.&lt;repository_suffix&gt;__ ).


In this mode you can create pages content like this:

1. Create a `gitea-pages` repo in any organization
2. Create a `gitea-pages` branch in this `gitea-pages` repository
3. Put your content in this branch. (eg index.html)

Your content will now be available on https://&lt;organization&gt;.&lt;your-server-domain&gt;/index.html


Beside this you can simply create a `gitea-pages` branch in any repository of the organization and push it to the gitea server. 
These pages are then available at: https://&lt;organization&gt;.&lt;your-server-domain&gt;/



Classic mode requires that you setup a _wildcard CNAME_ in DNS for your gitea pages host (see below for more details). You also need a _wildcard HTTPS_ certificate if you want to run with HTTPS.


# Usage

To run the server in __simple__ mode you need a minimal configuration (filename _Caddyfile_) like this (replace _your-gitea-server_ and _gitea_access_token_):

```Caddyfile
{
    order gitea-pages before file_server
}
:8080
gitea-pages {
    server https://your-gitea-server
    token gitea-access-token
}
```


Afterward you can simply run:

```bash
./caddy run --config Caddyfile
```

If you build/use the docker container you can run it like this:

```bash
# Create a Caddyfile (configuration file) first !
docker run -v $(pwd)/Caddyfile:/etc/caddy/Caddyfile -p 8080:8080 gitea-pages

# run the image hosted with this github user
docker run -v $(pwd)/Caddyfile:/etc/caddy/Caddyfile -p 8080:8080  ghcr.io/simongregorebner/gitea-pages:0.0.1
```

## Configuration

These are the possible configuration options with their defaults:

```Caddyfile
{
    order gitea-pages before file_server
}
:8080
gitea-pages {
    server https://your-gitea-server
    token gitea-access-token
    pages_branch gitea-pages
    postfix_pages_repository gitea-pages
    url_scheme simple
}
log {
    level debug
}
```

| Option | Description |
|----|----|
| server | The URL of your Gitea server.  |
| token | Your access token for the Gitea API. This token is used to authenticate requests made to the Gitea server.|
| pages_branch | The branch in your repository that contains the static files for your website or documentation. By default, this would be the branch "gitea-pages" |
| pages_repository | The default repository of an organization |
| postfix_pages_repository | The (domain) postfix used for the pages repository. __This setting is only used if pages_repository is not set!__ (This could be the domain where your site will be accessible, such as "gitea.io".) |
| url_scheme | The URL scheme to use for the pages. "simple" or "classic |


For example, if you want to assemble the same "look and feel" then on github you could set following settings in the config:

|||
|----|----|
| pages_branch | gh-pages |
| postfix_pages_repository | github.io |
| url_scheme | classic |


# Development

The project is implemented as an extension for [Caddy](https://github.com/caddyserver/caddy) .
More details can be found at: https://caddyserver.com/docs/extending-caddy

## Build
To build the server `xcaddy` need to be installed.

```bash
# Installing xcaddy
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

To build the server binary simply run

```bash
# Build a specific version
xcaddy build --with github.com/simongregorebner/gitea-pages@v0.0.1

# Build with specific commit (that was pushed to github)
xcaddy build --with github.com/simongregorebner/gitea-pages@f4a6a77

# Build locally
xcaddy build --with github.com/simongregorebner/gitea-pages=.
```

## Docker
To build the server and create a docker image you can use:

```bash
docker build -t gitea-pages .

# cross platform build
docker build --platform=linux/amd64 -t gitea-pages .

# example build image for this github user
docker build --platform=linux/amd64 -t ghcr.io/simongregorebner/gitea-pages:0.0.1 .
```

## Testing

If you are running/testing the server in __simple__ mode:
```bash
curl "http://localhost:8080/<organization>/<repository>[/index.html | /<path>]"
```

If running in __classic__ mode:
```bash
# testing organization repo
curl -H "Host: <organisation>.<your testserver name>" "http://localhost:8080/"
# testing a specific repo of the organization
curl -H "Host: <organisation>.<your testserver name>" "http://localhost:8080/<repository>"
```

# Acknowledgements

This project is an extremely simplified rewrite of the https://github.com/42wim/caddy-gitea project. 