# Overview

This is an all in one pages server for [Gitea](https://gitea.com). With default settings it works and uses similar/same conventions as [Github Pages](https://pages.github.com).

You can operate it in 2 modes, either __simple__ (default) or __classic__ (similar to how Github Pages operates).

In __simple__ mode no special DNS setup is required and the access to the hosted sides are always according to the pattern __http(s)://&lt;your-server-hostname&gt;/&lt;organization&gt;/&lt;repository&gt;__ 

In __classic__ mode the access to the pages goes according to these two patterns: __http(s)://&lt;organization&gt;.&lt;your-server-hostname&gt;/&lt;repository&gt;__ or __http(s)://&lt;organization&gt;.&lt;your-server-hostname&gt;__ (with default configuration this serves the content of the repo named __&lt;organization&gt;.github.io__ of the organization)

It is realized as a plugin for [Caddy v2](https://github.com/caddyserver/caddy).


# Acknowledgements

This project is an extremely simplifies rewrite of the https://github.com/42wim/caddy-gitea project. 