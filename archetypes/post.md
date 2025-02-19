---
title: {{ .Title }}
meta_title: {{ .MetaTitle }}
description: {{ .Description }}
date: {{ .Date }}
lastmod: {{ .Lastmod }}
image: {{ .Image }}
categories: {{ .Categories }}
author: {{ .Author }}
tags: {{ .Tags }}
draft: {{ .Draft }}
{{- with .Toc }}
toc: {{ . }}
{{- end }}
{{- with .Weight }}
weight: {{ . }}
{{- end }}
{{- with .Comments }}
comments: {{ . }}
{{- end }}
{{- with .Slug }}
slug: {{ . }}
{{- end }}
---

{{ .Content }} 